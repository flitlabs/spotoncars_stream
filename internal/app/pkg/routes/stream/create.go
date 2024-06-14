package stream

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/tokens"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

const (
	keyName      = "jobs"
	lockKey      = "lock:" + keyName
	lockValue    = "key:locked"
	lockDuration = 2 * time.Second
	timeout      = 5 * time.Second
)

type body struct {
	BookingID string `json:"booking_id" validate:"required,min=1"`
}

// Create is a route that is used to create a new stream for the given booking id
func create(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var reqBody body
	v := validator.New()

	if err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		lib.JSONResponse(w, http.StatusUnsupportedMediaType, "failed to decode the request body")
		return
	}

	if err := v.Struct(reqBody); err != nil {
		log.Error().Err(err).Msg("validation error, invalid data is provided")
		lib.JSONResponse(w, http.StatusBadRequest, "please provide a proper booking id")
		return
	}

	driverID := r.Context().Value(middlewares.DriverContextKey).(int)

	client := c.R.DB
	var available []int

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	serverBusyErr := fmt.Errorf("server is busy right now, no partitions are currently available")

	err := func() error {
		acquired, err := c.R.AcquireLock(ctx, client, lockKey, lockValue, lockDuration, 100*time.Millisecond)
		if err != nil {
			return err
		}
		if !acquired {
			return serverBusyErr
		}
		defer c.R.ReleaseLock(client, lockKey)

		payload := client.SMembers(r.Context(), keyName).Val()
		jobs := make(map[int]struct{})
		for _, data := range payload {
			job, err := strconv.Atoi(data)
			if err != nil {
				client.SRem(r.Context(), keyName, data)
				return err
			}
			jobs[job] = struct{}{}
		}
		partitions := make([]int, 10)
		for i := range partitions {
			partitions[i] = i
		}
		for _, partition := range partitions {
			_, found := jobs[partition]
			if !found {
				available = append(available, partition)
			}
		}

		if len(available) == 0 {
			return serverBusyErr
		}

		err = client.SAdd(r.Context(), keyName, available[0]).Err()
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		if errors.Is(err, serverBusyErr) {
			lib.JSONResponse(w, http.StatusConflict, "server is busy right now, please try again later")
			return
		}

		log.Error().Err(err)
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	partition := available[0]

	err = client.SetEx(r.Context(), reqBody.BookingID, partition, 24*time.Hour).Err()
	if err != nil {
		log.Error().Err(err).Msg("failed to assing the booking to the partition")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	bt := tokens.BookingToken{
		C: c,
		E: e,
	}
	token, err := bt.Create(r.Context(), strconv.Itoa(driverID), reqBody.BookingID, partition)
	if err != nil {
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong, please try again later")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "booking_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(e.BookingTokenExpires.Seconds()),
		Expires:  time.Now().UTC().Add(e.BookingTokenExpires).UTC(),
	})

	lib.JSONResponseWInterface(w, http.StatusOK, map[string]interface{}{
		"booking_token": token,
	})
}
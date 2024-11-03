package main_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/m1yon/starr-queue-cleaner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/starr"
	"golift.io/starr/sonarr"
)

func TestQueueCleaning(t *testing.T) {
	fakeServer := NewFakeSonarrServer(t)
	defer fakeServer.Close()

	c := starr.New("test-api-key", fakeServer.URL(), time.Second*5)
	sonarr := sonarr.New(c)

	queue, err := sonarr.GetQueue(100, 100)
	require.NoError(t, err)
	require.Equal(t, 2, queue.TotalRecords)

	err = main.CleanQueue(sonarr)
	require.NoError(t, err)

	queue, err = sonarr.GetQueue(100, 100)
	require.NoError(t, err)
	assert.Equal(t, 1, queue.TotalRecords)
	assert.Equal(t, int64(2), queue.Records[0].ID)
}

type FakeSonarrServer struct {
	server       *httptest.Server
	mux          *http.ServeMux
	queueRecords []*sonarr.QueueRecord
}

func NewFakeSonarrServer(t *testing.T) *FakeSonarrServer {
	t.Helper()
	fs := &FakeSonarrServer{
		mux: http.NewServeMux(),
		queueRecords: []*sonarr.QueueRecord{
			{ID: 1},
			{ID: 2},
		},
	}

	fs.mux.HandleFunc("GET /api/v3/queue", func(w http.ResponseWriter, r *http.Request) {
		result := sonarr.Queue{
			TotalRecords: len(fs.queueRecords),
			Records:      fs.queueRecords,
		}

		resp, err := json.Marshal(result)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})
	fs.mux.HandleFunc("DELETE /api/v3/queue/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		matchingIndex := -1
		for i, record := range fs.queueRecords {
			if record.ID == id {
				matchingIndex = i
				break
			}
		}
		if matchingIndex == -1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		fs.queueRecords = append(fs.queueRecords[:matchingIndex], fs.queueRecords[matchingIndex+1:]...)

		w.WriteHeader(http.StatusOK)
	})

	fs.server = httptest.NewServer(fs.mux)
	return fs
}

func (fs *FakeSonarrServer) Close() {
	fs.server.Close()
}

func (fs *FakeSonarrServer) URL() string {
	return fs.server.URL
}

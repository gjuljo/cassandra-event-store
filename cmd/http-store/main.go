package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"encoding/json"
	"io/ioutil"
	"my/esexample/store"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RemoteStorageHandler struct {
	EventStore store.EventStore
}

// HandleFindEventsByUUID ...
func (me *RemoteStorageHandler) HandleFindEventsByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	events, err := me.EventStore.Find(store.EventID(uuid))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := &store.FindEventsByTypeResult{
		Events: events,
	}

	if len(events) > 0 {
		result.Latest = events[len(events)-1].TimeStamp
	}

	c.JSON(http.StatusOK, result)
}

// HandleFindEventsByType ...
func (me *RemoteStorageHandler) HandleFindEventsByType(c *gin.Context) {
	stype := c.Param("type")
	since := c.Query("since")
	size := c.Query("size")

	// check whether version is an integer
	itype, err := strconv.Atoi(stype)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// check whether since is an integer
	isince, err := strconv.Atoi(since)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// check whether size is an integer
	isize, err := strconv.Atoi(size)

	if err != nil {
		isize = 100
	}

	events, latest, err := me.EventStore.GetEventsByType(store.EventType(itype), int64(isince), isize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := &store.FindEventsByTypeResult{
		Events: events,
		Latest: latest,
	}

	c.JSON(http.StatusOK, result)
}

// HandleUpdateEventByUUID ...
func (me *RemoteStorageHandler) HandleUpdateEventByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	sversion := c.Param("version")

	// check whether version is an integer
	iversion, err := strconv.Atoi(sversion)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jsondata, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// parse all input data
	var events []store.StoreEvent

	err = json.Unmarshal(jsondata, &events)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = me.EventStore.Update(store.EventID(uuid), iversion, events)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// Get the value of the environment variable key or the fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	log.Print("REMOTE EVENT-STORE")

	store, _ := store.NewCassandraEventStore(&store.CassandraEventStoreConfig{
		Hosts:       []string{"localhost"},
		Keyspace:    "eventstore",
		WriteQuorum: "QUORUM",
		ReadQuorum:  "LOCAL_QUORUM",
	})

	handler := &RemoteStorageHandler{EventStore: store}

	r := gin.New()
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.HEAD("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/events/:uuid", handler.HandleFindEventsByUUID)
	r.POST("/api/v1/events/:uuid/:version", handler.HandleUpdateEventByUUID)
	r.GET("/api/v1/types/:type", handler.HandleFindEventsByType)

	r.GET("/health/liveness", func(c *gin.Context) { c.JSON(http.StatusOK, "OK") })
	r.GET("/health/readiness", func(c *gin.Context) { c.JSON(http.StatusOK, "OK") })

	// Listen
	port := getEnv("PORT", "8091")
	r.Run(":" + port)
}

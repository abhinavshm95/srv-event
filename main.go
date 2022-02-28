package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"vh-srv-event/audience"
	"vh-srv-event/broadcasturl"
	"vh-srv-event/event"
	"vh-srv-event/item"
	part "vh-srv-event/participant"
	partoptn "vh-srv-event/partoptn"
	"vh-srv-event/partstatus"
	"vh-srv-event/platform"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

type Controllers struct {
	Participant         part.Participant
	ParticipationOption partoptn.ParticipationOption
	Platform            platform.Platform
	Audience            audience.Audience
	BroadcastURL        broadcasturl.BroadcastURL
	Item                item.Item
	ItemBroadcastURL    item.ItemBroadcastURL
	Event               event.Event
	EventItem           event.EventItem
	EventPartOption     event.EventPartOption
	ParticipationStatus partstatus.ParticipationStatus
}

// cfg is the struct type that contains fields that stores the necessary configuration
// gathered from the environment.
var cfg struct {
	DBUser   string `envconfig:"DB_USER" default:"postgres"`
	DBPass   string `envconfig:"DB_PASSWORD" default:"password"`
	DBName   string `envconfig:"DB_DATABASE" default:"event"`
	DBHost   string `envconfig:"DB_HOST" default:"localhost"`
	DBPort   string `envconfig:"DB_PORT" default:"5432"`
	APP_PORT string `envconfig:"APP_PORT" default:"8080"`
}

type Router struct {
	server              *gin.Engine
	participant         part.Participant
	participationOption partoptn.ParticipationOption
	platform            platform.Platform
	audience            audience.Audience
	broadcastURL        broadcasturl.BroadcastURL
	item                item.Item
	itemBroadcastURL    item.ItemBroadcastURL
	event               event.Event
	eventItem           event.EventItem
	eventPartOption     event.EventPartOption
	participationStatus partstatus.ParticipationStatus
}

func NewRouter(server *gin.Engine, controller Controllers) *Router {
	return &Router{
		server,
		controller.Participant,
		controller.ParticipationOption,
		controller.Platform,
		controller.Audience,
		controller.BroadcastURL,
		controller.Item,
		controller.ItemBroadcastURL,
		controller.Event,
		controller.EventItem,
		controller.EventPartOption,
		controller.ParticipationStatus,
	}
}
func (r *Router) Init() {

	basePath := r.server.Group("/v1")

	participant := basePath.Group("/participant")
	{
		participant.POST("/", r.participant.CreateNewParticipant)
		participant.PATCH("/:id", r.participant.UpdateParticipantByID)
		participant.DELETE("/:id", r.participant.DeleteParticipantByID)
		participant.GET("/:id", r.participant.GetParticipantById)
		participant.GET("email/:email", r.participant.GetParticipantByEmail)
		participant.GET("keycloakid/:id", r.participant.GetParticipantByKeycloakID)
	}
	basePath.GET("/participants", r.participant.GetAllParticipant)

	participationOption := basePath.Group("/participation-option")
	{
		participationOption.POST("/", r.participationOption.CreateNewParticipationOption)
		participationOption.PATCH("/:name", r.participationOption.UpdateParticipationOptionByName)
		participationOption.DELETE("/:name", r.participationOption.DeleteParticipationOptionByName)
		participationOption.GET("/:name", r.participationOption.GetParticipationOptionByName)
	}
	basePath.GET("/participation-options", r.participationOption.GetAllParticipationOption)

	platform := basePath.Group("/platform")
	{
		platform.POST("/", r.platform.CreateNewPlatform)
		platform.PATCH("/:name", r.platform.UpdatePlatformByName)
		platform.DELETE("/:name", r.platform.DeletePlatformByName)
		platform.GET("/:name", r.platform.GetPlatformByName)
	}
	basePath.GET("/platforms", r.platform.GetAllPlatform)

	audience := basePath.Group("/audience")
	{
		audience.POST("/", r.audience.CreateNewAudience)
		audience.PATCH("/:name", r.audience.UpdateAudienceByName)
		audience.DELETE("/:name", r.audience.DeleteAudienceByName)
		audience.GET("/:name", r.audience.GetAudienceByName)
	}
	basePath.GET("/audiences", r.audience.GetAllAudience)

	broadcastURL := basePath.Group("/broadcasturl")
	{
		broadcastURL.POST("/", r.broadcastURL.CreateNewBroadcastURL)
		broadcastURL.PATCH("/:id", r.broadcastURL.UpdateBroadcastURLByID)
		broadcastURL.DELETE("/:id", r.broadcastURL.DeleteBroadcastURLByID)
		broadcastURL.GET("/:id", r.broadcastURL.GetBroadcastURLByID)
	}
	basePath.GET("/broadcasturls", r.broadcastURL.GetAllBroadcastURL)

	item := basePath.Group("/item")
	{
		item.POST("/", r.item.CreateNewItem)
		item.GET("/:id", r.item.GetItemByID)
		item.PATCH("/:id", r.item.UpdateItemByID)
		item.DELETE("/:id", r.item.DeleteItemByID)
	}
	basePath.GET("/items", r.item.GetAllItem)

	itemBroadcastUrl := basePath.Group("/item-broadcasturl")
	{
		itemBroadcastUrl.POST("/", r.itemBroadcastURL.CreateNewItemBroadcastURL)
		itemBroadcastUrl.GET("/:id", r.itemBroadcastURL.GetItemBroadcastURLByID)
		itemBroadcastUrl.PATCH("/:id", r.itemBroadcastURL.UpdateItemBroadcastURLByID)
		itemBroadcastUrl.DELETE("/:id", r.itemBroadcastURL.DeleteItemBroadcastURLByID)
	}
	basePath.GET("/item-broadcasturls", r.itemBroadcastURL.GetAllItemBroadcastURL)

	event := basePath.Group("/event")
	{
		event.POST("/", r.event.CreateNewEvent)
		event.GET("/:id", r.event.GetEventByID)
		event.PATCH("/:id", r.event.UpdateEventByID)
		event.DELETE("/:id", r.event.DeleteEventByID)
		event.DELETE("/hard/:id", r.event.DeleteHardEventByID)
	}
	basePath.GET("/events", r.event.GetAllEvent)

	eventItem := basePath.Group("/event-item")
	{
		eventItem.POST("/", r.eventItem.CreateNewEventItem)
		eventItem.GET("/:id", r.eventItem.GetEventItemByID)
		eventItem.PATCH("/:id", r.eventItem.UpdateEventItemByID)
		eventItem.DELETE("/:id", r.eventItem.DeleteEventItemByID)
	}
	basePath.GET("/event-items", r.eventItem.GetAllEventItem)

	eventPartOption := basePath.Group("/event-part-option")
	{
		eventPartOption.POST("/", r.eventPartOption.CreateNewEventPartOption)
		eventPartOption.GET("/:id", r.eventPartOption.GetEventPartOptionByID)
		eventPartOption.PATCH("/:id", r.eventPartOption.UpdateEventPartOptionByID)
		eventPartOption.DELETE("/:id", r.eventPartOption.DeleteEventPartOptionByID)
	}
	basePath.GET("/event-part-options", r.eventPartOption.GetAllEventPartOption)

	participationStatus := basePath.Group("/participation-status")
	{
		participationStatus.POST("/", r.participationStatus.CreateNewParticipationStatus)
		participationStatus.GET("/:id", r.participationStatus.GetParticipationStatusByID)
		participationStatus.PATCH("/:id", r.participationStatus.UpdateParticipationStatusByID)
		participationStatus.DELETE("/:id", r.participationStatus.DeleteParticipationStatusByID)
	}
	basePath.GET("/participation-statuses", r.participationStatus.GetAllParticipationStatus)
}

func main() {
	route := gin.Default()

	if err := envconfig.Process("LIST", &cfg); err != nil {
		log.Fatalln("Error while fetching env file")
		return
	}

	databaseURL := "postgres://" + cfg.DBUser + ":" + cfg.DBPass + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		fmt.Fprintf(os.Stderr, "Connection url: %s", databaseURL)
		os.Exit(1)
	}
	defer conn.Close()

	participant := part.NewParticipant(conn)
	participationOption := partoptn.NewParticipationOption(conn)
	platform := platform.NewPlatform(conn)
	audience := audience.NewAudience(conn)
	broadcasturl := broadcasturl.NewBroadcastURL(conn)
	itemBroadcastURL := item.NewItemBroadcastURL(conn)
	item := item.NewItem(conn)
	eventPartOption := event.NewEventPartOption(conn)
	eventItem := event.NewEventItem(conn)
	event := event.NewEvent(conn)
	participationStatus := partstatus.NewParticipationStatus(conn)

	r := NewRouter(route, Controllers{
		Participant:         participant,
		ParticipationOption: participationOption,
		Platform:            platform,
		Audience:            audience,
		BroadcastURL:        broadcasturl,
		Item:                item,
		ItemBroadcastURL:    itemBroadcastURL,
		Event:               event,
		EventItem:           eventItem,
		EventPartOption:     eventPartOption,
		ParticipationStatus: participationStatus,
	})

	r.Init()

	route.Run("localhost:" + cfg.APP_PORT)
}

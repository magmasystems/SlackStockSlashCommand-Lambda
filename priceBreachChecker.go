package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ev "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"github.com/magmasystems/SlackStockSlashCommand/alerts"
	config "github.com/magmasystems/SlackStockSlashCommand/configuration"
	"github.com/magmasystems/SlackStockSlashCommand/slackmessaging"
	"github.com/magmasystems/SlackStockSlashCommand/stockbot"
)

var theBot *stockbot.Stockbot
var theAlertManager *alerts.AlertManager
var appSettings *config.AppSettings

func init() {
	configMgr := new(config.ConfigManager)
	appSettings = configMgr.Config()

	theBot = stockbot.CreateStockbot()
	// defer theBot.Close()

	// Create the AlertManager
	theAlertManager = alerts.CreateAlertManager(theBot)
	// defer theAlertManager.Dispose()
}

func main() {
	lambda.Start(priceBreachChecker)
}

func priceBreachChecker(ctx context.Context, event ev.CloudWatchEvent) (int, error) {
	lambdaContext, _ := lambdacontext.FromContext(ctx)
	log.Printf("In priceBreachChecker handler: context is %+v\n", lambdaContext)
	log.Printf("In priceBreachChecker handler: event is %+v\n", event)

	checkForPriceBreaches()

	return 0, nil
}

// checkForPriceBreaches - checks for price breaches
func checkForPriceBreaches() {
	fmt.Println("checkForPriceBreaches: Checking for price breaches at " + time.Now().String())

	theAlertManager.CheckForPriceBreaches(theBot, func(notification alerts.PriceBreachNotification) {
		log.Println("The notification to Slack is:")
		log.Println(notification)
		outputText := fmt.Sprintf("%s has gone %s the target price of %3.2f. The current price is %3.2f.\n",
			notification.Symbol, notification.Direction, notification.TargetPrice, notification.CurrentPrice)

		slackmessaging.PostSlackNotification(notification.SlackUserName, notification.Channel, outputText, appSettings)
	})

	fmt.Printf("checkForPriceBreaches: Finished checking for price breaches at %s\n", time.Now().String())
}

package main

import (
	"context"
	"log"

	fyneapp "fyne.io/fyne/v2/app"
	appservice "github.com/blackscorp/albion-helper/internal/app"
	"github.com/blackscorp/albion-helper/internal/catalog"
	"github.com/blackscorp/albion-helper/internal/marketapi"
	"github.com/blackscorp/albion-helper/internal/storage"
	"github.com/blackscorp/albion-helper/internal/ui"
)

func main() {
	dbPath, err := storage.DefaultDatabasePath()
	if err != nil {
		log.Fatal(err)
	}
	repository, err := storage.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer repository.Close()
	prices, err := repository.Prices(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	a := fyneapp.NewWithID("de.blackscorp.albion-helper")
	syncService := appservice.SyncService{Store: repository}
	w := ui.NewWindowWithData(a, prices, func(server string, itemIDs []string, markets []catalog.Market) ([]catalog.Price, error) {
		return syncService.SyncAt(context.Background(), marketapi.Server(server), itemIDs, markets)
	})
	w.ShowAndRun()
}

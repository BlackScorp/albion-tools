package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	fyneapp "fyne.io/fyne/v2/app"
	appservice "github.com/blackscorp/albion-helper/internal/app"
	"github.com/blackscorp/albion-helper/internal/catalog"
	"github.com/blackscorp/albion-helper/internal/marketapi"
	"github.com/blackscorp/albion-helper/internal/storage"
	"github.com/blackscorp/albion-helper/internal/ui"
)

func main() {
	logFile, logPath := openLogFile()
	if logFile != nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Albion Helper gestartet; Logdatei: %s", logPath)

	dbPath, err := storage.DefaultDatabasePath()
	if err != nil {
		log.Printf("Datenbankpfad konnte nicht bestimmt werden: %v", err)
		log.Fatal(err)
	}
	repository, err := storage.Open(dbPath)
	if err != nil {
		log.Printf("Datenbank konnte nicht geöffnet werden: %v", err)
		log.Fatal(err)
	}
	defer repository.Close()
	log.Printf("Datenbank geöffnet: %s", dbPath)
	prices, err := repository.Prices(context.Background())
	if err != nil {
		log.Printf("Preise konnten nicht geladen werden: %v", err)
		log.Fatal(err)
	}
	log.Printf("Lokale Marktbeobachtungen geladen: %d", len(prices))
	a := fyneapp.NewWithID("de.blackscorp.albion-helper")
	syncService := appservice.SyncService{Store: repository}
	w := ui.NewWindowWithData(a, prices, func(server string, itemIDs []string, markets []catalog.Market) ([]catalog.Price, error) {
		return syncService.SyncAt(context.Background(), marketapi.Server(server), itemIDs, markets)
	})
	w.ShowAndRun()
}

func openLogFile() (*os.File, string) {
	path, err := os.Executable()
	if err == nil {
		logPath := filepath.Join(filepath.Dir(path), "albion-helper.log")
		if file, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); openErr == nil {
			return file, logPath
		}
	}
	if configDir, configErr := os.UserConfigDir(); configErr == nil {
		logPath := filepath.Join(configDir, "Albion Helper", "albion-helper.log")
		if mkdirErr := os.MkdirAll(filepath.Dir(logPath), 0o755); mkdirErr == nil {
			if file, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); openErr == nil {
				return file, logPath
			}
		}
	}
	return nil, "stderr (Datei konnte nicht erstellt werden)"
}

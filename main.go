package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"www-apps.univ-lehavre.fr/forge/themd5destroyers/theleaddestroyer/adapters/websocket"
)

func main() {
	// Initialisez le client Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Adresse du serveur Redis
		Password: "",           // Pas de mot de passe
		DB:       0,            // Base de données par défaut
	})

	// Testez la connexion Redis
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Échec de la connexion à Redis: %v\n", err)
	}
	log.Printf("Connecté à Redis: %s\n", pong)

	// Créez un nouveau ConnectionManager avec le client Redis
	cm := websocket.NewConnectionManager(rdb)

	// Configurez les routes HTTP
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Nouvelle demande de connexion WebSocket")
		websocket.HandleWebSocket(cm, w, r)
	})

	// Configurez le serveur
	port := "8080"
	server := &http.Server{
		Addr:    "0.0.0.0:8080", // Écoute sur toutes les interfaces réseau
		Handler: nil,
	}

	// Démarrez le serveur dans une goroutine
	go func() {
		log.Printf("Démarrage du serveur WebSocket à ws://localhost:%s/ws\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur lors du démarrage du serveur: %v\n", err)
		}
	}()

	// Gardez le serveur en fonctionnement indéfiniment
	log.Println("Le serveur est en cours d'exécution. Appuyez sur CTRL+C pour arrêter.")
	select {} // Bloque indéfiniment
}

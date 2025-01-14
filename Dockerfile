# Utilisez une image de base Go
FROM golang:1.20-alpine AS builder

# Définissez le répertoire de travail
WORKDIR /app

# Copiez les fichiers nécessaires
COPY . .

# Téléchargez les dépendances
RUN go mod download

# Compilez l'application
RUN go build -o hash_extractor .

# Utilisez une image Alpine légère pour l'exécution
FROM alpine:latest

# Copiez le binaire compilé
COPY --from=builder /app/hash_extractor /hash_extractor

# Commande pour exécuter l'application
CMD ["/hash_extractor", "s", "ws://backend:8080/ws"]
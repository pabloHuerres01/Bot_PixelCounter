package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Declara la variable 'dg' aquí
var dg *discordgo.Session
var token string // Variable para almacenar el token del bot

func main() {
	// Intenta cargar las variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error al cargar el archivo .env:", err)
		return // Salir si hay un error
	}

	// Acceder a una variable de entorno
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("El token no está definido.")
		return
	}

	// Crear la sesión
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creando la sesión:", err)
		return
	}

	// Abrir la conexión
	err = dg.Open()
	if err != nil {
		fmt.Println("Error abriendo la conexión:", err)
		return
	}

	fmt.Println("Bot conectado y funcionando.")

	// Mantener el bot en ejecución
	select {} // Este select vacío mantiene el bot en ejecución
}

// Manejador de mensajes
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	print("pasa-")
	// Ignorar mensajes del propio bot
	if m.Author.ID == s.State.User.ID {
		return
	}
	print("pasa1")
	// Responder a un mensaje específico
	if m.Content == "!hola" {
		print("pasa2")
		s.ChannelMessageSend(m.ChannelID, "¡Hola Mundo!")
	}
}

package main

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/otiai10/gosseract/v2"
	"github.com/sashabaranov/go-openai"
)

var (
	dg     *discordgo.Session
	token  string
	client *openai.Client
)

func main() {
	fmt.Println("Cargando variables de entorno...")
	// Cargar variables de entorno
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error al cargar el archivo .env:", err)
		return
	}

	token = os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("El token no está definido.")
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("La API Key de OpenAI no está definida.")
		return
	}

	client = openai.NewClient(apiKey)

	fmt.Println("Creando sesión de Discord...")
	// Crear la sesión
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creando la sesión:", err)
		return
	}

	// Añadir el manejador de mensajes
	dg.AddHandler(messageCreate)

	fmt.Println("Abriendo conexión...")
	// Abrir la conexión
	err = dg.Open()
	if err != nil {
		fmt.Println("Error abriendo la conexión:", err)
		return
	}

	fmt.Println("Bot conectado y funcionando.")
	select {} // Mantener el bot en ejecución
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Manejando mensaje de:", m.Author.Username)

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Procesar imágenes
	for _, attachment := range m.Attachments {
		fmt.Printf("Procesando archivo adjunto: %s\n", attachment.URL)
		if attachment.Width > 0 && attachment.Height > 0 { // Asegurarse de que es una imagen
			number := extractNumberFromImage(attachment.URL)
			fmt.Printf("Número extraído: %s\n", number)

			response := askGPT(number) // Preguntar a GPT
			s.ChannelMessageSend(m.ChannelID, response)
			return
		}
	}

	fmt.Println("No se encontraron archivos adjuntos válidos.")
}

// Extraer número de la imagen usando OCR
func extractNumberFromImage(imageURL string) string {
	fmt.Printf("Entrando en extractNumberFromImage con URL: %s\n", imageURL)

	client := gosseract.NewClient()
	defer client.Close()

	// Descargar la imagen
	resp, err := http.Get(imageURL)
	if err != nil {
		fmt.Println("Error al descargar la imagen:", err)
		return "Error al procesar la imagen"
	}
	defer resp.Body.Close()

	// Leer la imagen
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Println("Error al decodificar la imagen:", err)
		return "Error al procesar la imagen"
	}

	// Guardar la imagen en un archivo temporal
	tempFile, err := ioutil.TempFile("", "temp_image_*.jpg")
	if err != nil {
		fmt.Println("Error al crear archivo temporal:", err)
		return "Error al procesar la imagen"
	}
	defer os.Remove(tempFile.Name()) // Asegurarse de que se elimine el archivo temporal

	// Codificar la imagen JPEG y guardarla en el archivo temporal
	if err := jpeg.Encode(tempFile, img, nil); err != nil {
		fmt.Println("Error al guardar la imagen temporal:", err)
		return "Error al procesar la imagen"
	}

	// Establecer la imagen en el cliente de gosseract
	if err := client.SetImage(tempFile.Name()); err != nil {
		fmt.Println("Error al establecer la imagen en Gosseract:", err)
		return "Error al procesar la imagen"
	}

	text, err := client.Text()
	if err != nil {
		fmt.Println("Error al obtener texto de la imagen:", err)
		return "No se pudo leer la imagen"
	}

	fmt.Println("Texto extraído de la imagen:", text) // Depuración

	// Buscar el primer número en el texto
	re := regexp.MustCompile(`\d+`)
	number := re.FindString(text)

	fmt.Println("Número extraído:", number) // Depuración

	if number == "" {
		return "No se encontró ningún número"
	}
	return number
}

// Función para preguntar a GPT
func askGPT(number string) string {
	fmt.Printf("Preguntando a GPT con el número: %s\n", number)

	prompt := fmt.Sprintf("¿Cuál es la interpretación del número %s?", number)
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})

	if err != nil {
		fmt.Println("Error al preguntar a GPT:", err)
		return "Error al preguntar a GPT"
	}

	return resp.Choices[0].Message.Content
}

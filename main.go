package main

import (
	"flag"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"gopkg.in/zeromq/goczmq.v4"
	"log"
	"os"
	"time"
)

type inputArgs struct {
	clientIp           string
	clientPort         int
	appPort            int
	audioInput         string
	audioInputFilePath string
	audioSineFreq      float64
	audioOutput        string
	audioOutputFilePath string
}

func holdOnMsg(router *goczmq.Sock) {
	for {
		input, err := router.RecvMessage()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		fmt.Printf("\nRecieved message from peer: %s\nMessage: ", input)
	}

}

func setupArguments () inputArgs {
	clientIp := flag.String("clientip", "127.0.0.1", "opposite address IP")
	clientPort := flag.Int("clientport", 5555, "port to connect to")
	appPort := flag.Int("port", 5555, "port to receive on")

	audioInput := flag.String("input", "sine", "Select input type, options include sine, jack, and file")
	audioFilePath := flag.String("inputFilePath", "./meme.mp3", "Optional input file path for program")
	audioSineFreq := flag.Float64("inputSineFreq", 1000., "Frequency of produced sine wave")

	audioOutput := flag.String("output", "jack", "select output type, options include jack and file")
	audioOutputFilePath := flag.String("outputFilePath", "output.opus", "output file path and type")

	flag.Parse()

	arguments := inputArgs{
		clientIp:           *clientIp,
		clientPort:         *clientPort,
		appPort:            *appPort,
		audioInput:         *audioInput,
		audioInputFilePath: *audioFilePath,
		audioSineFreq:      *audioSineFreq,
		audioOutput:        *audioOutput,
		audioOutputFilePath: *audioOutputFilePath,
	}
	return arguments
}

func createRouter (appPort *int) *goczmq.Sock {
	routerString := fmt.Sprintf("tcp://*:%d",*appPort)
	router, err := goczmq.NewRouter(routerString)
	if err != nil {
		log.Fatal(err)
	}

	defer router.Destroy()

	log.Println("router created and bound on port %d", *appPort)

	return router
}

func createDealer (clientIp *string, clientPort *int) *goczmq.Sock {
	dealerEndpoint := fmt.Sprintf("tcp://%s:%d", *clientIp, *clientPort)
	dealer, err := goczmq.NewDealer(dealerEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer dealer.Destroy()

	log.Println("dealer created and connected on port %d", *clientPort)

	err = dealer.SendFrame([]byte("You have connected successfully"), 0)

	return  dealer

}

func main() {
	arguments := setupArguments()

	var err error


	//if *sendAudio == "Yes" {
	//	router := createRouter(appPort)
	//	dealer := createDealer(clientIp, clientPort)
	//
	//	go holdOnMsg(router)
	//
	//	reader := bufio.NewReader(os.Stdin)
	//	input := ""
	//
	//	for (input != "Q") {
	//		//fmt.Print("Message: ")
	//		input, err = reader.ReadString('\n')
	//		fmt.Print("Message: ")
	//
	//		dealer.SendFrame([]byte(input), 0)
	//	}
	//
	//}

	portaudio.Initialize()
	defer portaudio.Terminate()

	inputCallback, err := getInputCallback(arguments.audioInput, arguments)

	//inputCallback := generateSineCallback(1e3, 48e3)
	//inputCallback := generateMp3FileCallback("./meme.mp3")

	stream, err := setupJackDevice(inputCallback)
	defer stream.Close()
	err = stream.Start()

	time.Sleep(time.Minute)

	print(stream, err)

	defer func() {
		fmt.Print("\nNightmare")
	}()

	if err != nil {
		fmt.Printf("Error detected at program termination: ", err)
	}

}

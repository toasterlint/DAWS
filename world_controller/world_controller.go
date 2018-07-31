package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"

	. "image"

	"github.com/Azure/azure-service-bus-go"
	"github.com/Pallinder/go-randomdata"
	"github.com/gorilla/mux"
	dao "github.com/toasterlint/DAWS/common/dao"
	models "github.com/toasterlint/DAWS/common/models"
	. "github.com/toasterlint/DAWS/common/utils"
)

var connStr string
var conn *servicebus.Namespace
var worldQueue *servicebus.Queue
var topic *servicebus.Topic
var sub *servicebus.Subscription
var runTrigger bool
var controllers []models.Controller
var settings models.Settings
var commonDAO = dao.DAO{Database: "world.db"}
var numCities = 0
var numBuildings = 0
var numPeople = 0
var looper = 0

func startHTTPServer() {
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("./html")))
	r.HandleFunc("/api/status", apiStatus).Methods("GET")
	//	r.HandleFunc("/api/triggerNext", apiTrigger).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func connectQueues() {
	var err error
	connStr := "Endpoint=sb://dawsim.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=CzhNWYYYcDHZjnWsVH/Psg4YnP9JQkei2IGR11EkkY8="
	conn, err = servicebus.NewNamespace(servicebus.NamespaceWithConnectionString(connStr))
	FailOnError(err, "Failed to connect to ServiceBus")

	worldQueue, err = conn.NewQueue("worldqueue")
	FailOnError(err, "Failed to Open worldqueue")
	topic, _ = conn.NewTopic(context.Background(), "worldbroadcast")
	sub, _ = topic.NewSubscription(context.Background(), "testsub1")
	Logger.Printf("Sub: %s", sub.Name)
}

func apiStatus(w http.ResponseWriter, r *http.Request) {
	LogToConsole("API Call made: status")
	w.Write([]byte("API Call made: status"))
}

//func apiTrigger(w http.ResponseWriter, r *http.Request) {
//	cityids, err := commonDAO.GetAllCityIDs()
//	FailOnError(err, "Failed to get city IDs")
//	msg := &models.WorldTrafficQueueMessage{WorldSettings: settings}
//	triggerNext(cityids, msg)
//	LogToConsole("Manually Trigger")
//	w.Write([]byte("Manually triggered"))
//}

func triggerNext(cities []uuid.UUID, worldtrafficmessage *models.WorldTrafficQueueMessage) {
	//tempMsgJSON, _ := json.Marshal(worldtrafficmessage)
	//	err := ch.Publish(
	//		"",                 // exchange
	//		worldtrafficq.Name, // routing key
	//		false,              // mandatory
	//		false,
	//		amqp.Publishing{
	//			DeliveryMode: amqp.Persistent,
	//			ContentType:  "application/json",
	//			Body:         []byte(tempMsgJSON),
	//		})
	//	FailOnError(err, "Failed to post to World Traffic Queue")
	//	for _, element := range cities {
	//		tempMsg := &models.WorldCityQueueMessage{WorldSettings: settings, City: element.ID.Hex()}
	//		tempMsgJSON, _ := json.Marshal(tempMsg)
	//		err := ch.Publish(
	//			"",              // exchange
	//			worldcityq.Name, // routing key
	//			false,           // mandatory
	//			false,
	//			amqp.Publishing{
	//				DeliveryMode: amqp.Persistent,
	//				ContentType:  "application/json",
	//				Body:         []byte(tempMsgJSON),
	//			})
	//		FailOnError(err, "Failed to post to World City Queue")
	//	}
}

func processTrigger() {
	realLastTime := time.Now()
	go printStatus()
	for runTrigger {
		time.Sleep(time.Microsecond * 500)
		message := servicebus.Message{Data: []byte("This is a test!!!!!!!!!!!!")}
		worldQueue.Send(context.Background(), &message)
		err := topic.Send(context.Background(), &message)
		FailOnError(err, "Failed to send message to topic")
		// first check if all controllers are ready (and that we have any)
		if len(controllers) == 0 {
			LogToConsole("No controllers")
			time.Sleep(time.Second * 5)
			continue
		}
		readyt := true
		readyc := true
		totalt := 0
		totalc := 0
		for i := range controllers {
			if controllers[i].Type == "traffic" {
				totalt++
			} else {
				totalc++
			}
			if controllers[i].Ready == false {
				if controllers[i].Type == "traffic" {
					readyt = false
				} else {
					readyc = false
				}
				break
			}
		}
		if readyt == false || readyc == false {
			continue
		}
		if totalt == 0 || totalc == 0 {
			continue
		}
		// make sure we don't go over max speed limit
		t := time.Now()
		dur := t.Sub(realLastTime)
		for i := range controllers {
			controllers[i].Ready = false
		}
		if dur > time.Duration(settings.WorldSpeed)*time.Millisecond {
			LogToConsole("Warning: world processing too slow, last duration was - " + dur.String())
		}
		//		cityids, err := commonDAO.GetAllCityIDs()
		//		FailOnError(err, "Failed to get city IDs")
		//		msg := &models.WorldTrafficQueueMessage{WorldSettings: settings}
		//		triggerNext(cityids, msg)
		settings.LastTime = settings.LastTime.Add(time.Second * 1)
		realLastTime = time.Now()
	}
}

func processMsgs() {
	//	for d := range msgs {
	//		tempController := models.Controller{}
	//		json.Unmarshal(d.Body, &tempController)
	//		found := false
	//		var tempRemove int
	//		for i := range controllers {
	//			if controllers[i].ID == tempController.ID {
	//				found = true
	//				controllers[i].Ready = tempController.Ready
	//				tempRemove = i
	//				break
	//			}
	//		}
	//		//LogToConsole("Controller found stats: " + strconv.FormatBool(found))
	//		if found == false {
	//			controllers = append(controllers, tempController)
	//		}
	//		// Remove the controller if it sent exit true
	//		if tempController.Exit == true {
	//			controllers = append(controllers[:tempRemove], controllers[tempRemove+1:]...)
	//		}
	//		//LogToConsole("Done")
	//		d.Ack(false)
	//	}
	listener, err := sub.Receive(context.Background(), func(ctx context.Context, msg *servicebus.Message) servicebus.DispositionAction {
		Logger.Println(string(msg.Data))
		return msg.Complete()
	})
	FailOnError(err, "Failed to start listener")
	defer listener.Close(context.Background())

	// Loop main thread
	forever := make(chan bool)
	<-forever
}

func runConsole() {
	// setup terminal
	reader := bufio.NewReader(os.Stdin)
ReadCommand:
	Logger.Print("Command: ")
	text, _ := reader.ReadString('\n')
	text = strings.Trim(text, "\n")
	switch text {
	case "exit":
		Logger.Print("Purging queues")
		//		_, err := ch.QueuePurge(worldcityq.Name, false)
		//		FailOnError(err, "Failed to purge World City Queue")
		//		_, err = ch.QueuePurge(worldtrafficq.Name, false)
		//		FailOnError(err, "Failed to purge World Traffic Queue")
		//		_, err = ch.QueuePurge(worldq.Name, false)
		//		FailOnError(err, "Failed to purge World Queue")
		Logger.Println("Saving settings...")
		//		err = commonDAO.SaveSettings(settings)
		//		FailOnError(err, "Failed to save settings")
		Logger.Println("Exiting...")
		os.Exit(0)
	case "status":
		if runTrigger {
			Logger.Println("Running...")
		} else {
			Logger.Println("Stopped...")
		}

		tcontrollers := 0
		ccontrollers := 0
		for i := range controllers {
			if controllers[i].Type == "traffic" {
				tcontrollers++
			} else {
				ccontrollers++
			}
		}
		Logger.Printf("Traffic Controllers: %d", tcontrollers)
		Logger.Printf("City Controllers: %d", ccontrollers)
		Logger.Printf("Current Real Time: %s", time.Now().Format("2006-01-02 15:04:05"))
		Logger.Printf("Current Simulated Time: %s", settings.LastTime.Format("2006-01-02 15:04:05"))
	case "start":
		runTrigger = true
		go processTrigger()
		Logger.Println("World simulation started")
	case "stop":
		runTrigger = false
		Logger.Println("World simulation stopped")
	case "help":
		fallthrough
	default:
		Logger.Println("Help: ")
		Logger.Println("   status - Check the status of the world")
		Logger.Println("   start - Start world simulation")
		Logger.Println("   stop - Stop world simulation")
		Logger.Println("   exit - Exit the App")
	}
	goto ReadCommand
}

func loadConfig() {
	var err error
	settings = models.Settings{}
	//settings, err = commonDAO.LoadSettings()
	FailOnError(err, "Failed to load settings")
	if settings.LastTime != (time.Time{}) {
		sett, _ := json.Marshal(settings)
		Logger.Println(string(sett))
	} else {
		LogToConsole("No settings found, creating defaults")
		var tempSettings models.Settings
		tempSettings.CarAccidentFatalityRate = 0.0001159
		tempSettings.ID, _ = uuid.NewV4()
		tempSettings.LastTime = time.Now()
		tempSettings.MurderRate = 0.000053
		tempSettings.ViolentCrimeRate = 0.00381
		tempSettings.WorldSpeed = 5000
		var speeds = []models.SpeedLimit{}
		var citySpeed = models.SpeedLimit{Location: "city", Value: 35}
		var noncitySpeed = models.SpeedLimit{Location: "noncity", Value: 70}
		speeds = append(speeds, citySpeed)
		speeds = append(speeds, noncitySpeed)
		tempSettings.SpeedLimits = speeds
		tempSettings.Diseases = []models.Disease{}
		//err := commonDAO.InsertSettings(tempSettings)
		settings = tempSettings
		FailOnError(err, "Failed to insert settings")
	}
	getBuildingsCount()
	getCitiesCount()
	getPeopleCount()
}

func printStatus() {
	for runTrigger {
		if runTrigger {
			Logger.Println("Running...")
		} else {
			Logger.Println("Stopped...")
		}

		tcontrollers := 0
		ccontrollers := 0
		for i := range controllers {
			if controllers[i].Type == "traffic" {
				tcontrollers++
			} else {
				ccontrollers++
			}
		}
		Logger.Printf("Traffic Controllers: %d", tcontrollers)
		Logger.Printf("City Controllers: %d", ccontrollers)
		Logger.Printf("Current Real Time: %s", time.Now().Format("2006-01-02 15:04:05"))
		Logger.Printf("Current Simulated Time: %s", settings.LastTime.Format("2006-01-02 15:04:05"))
		time.Sleep(time.Second * 10)
	}
}

func main() {
	commonDAO.Open()
	defer commonDAO.Close()
	// Set some initial variables
	InitLogger()
	loadConfig()
	controllers = []models.Controller{}

	//init rabbit
	connectQueues()
	go processMsgs()

	// Start Web Server
	go startHTTPServer()

	// Start Console
	go runConsole()

	// Check to see if fresh world
	if numBuildings == 0 || numCities == 0 || numPeople == 0 {
		aWholeNewWorld()
	}

	// start world simulation
	runTrigger = true
	go processTrigger()

	Logger.Println("done initializing")

	// Loop main thread
	forever := make(chan bool)
	<-forever
}

func getCitiesCount() {
	//	numCities, _ = commonDAO.GetCitiesCount()
	//	Logger.Printf("Number of Cities: %d", numCities)
}

func getPeopleCount() {
	//	numPeople, _ = commonDAO.GetPeopleCount()
	//	Logger.Printf("Nummber of People in world: %d", numPeople)
}

func getBuildingsCount() {
	//	numBuildings, _ = commonDAO.GetBuildingsCount()
	//	Logger.Printf("Nummber of Buildings in world: %d", numBuildings)
}

func aWholeNewWorld() {
	LogToConsole("Starting a whole new world... don't you dare close your eyes!")
	// Create a new city somewhere random in the world that is 1 sq mile
	newCity := models.City{}
	newCity.ID, _ = uuid.NewV4()

	cityNameGenURL := "https://www.mithrilandmages.com/utilities/CityNamesServer.php?count=1&dataset=united_states&_=1531715721885"
	req, err := http.NewRequest("GET", cityNameGenURL, nil)
	FailOnError(err, "Failed on http.NewRequest")
	webClient := &http.Client{}
	resp, err := webClient.Do(req)
	FailOnError(err, "Failed to get name")
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		FailOnError(err2, "Failed to read html body")
		newCity.Name = string(bodyBytes)
	}
	newCity.TopLeft = Point{X: randomdata.Number(5274720), Y: randomdata.Number(5274720)}
	newCity.BottomRight = Point{X: newCity.TopLeft.X + 5280, Y: newCity.TopLeft.Y + 5280}
	newCity.Established = settings.LastTime
	err = commonDAO.CreateCity(newCity)
	FailOnError(err, "Failed to create new city")
	Logger.Printf("Created City: %s", newCity.Name)

	canWeFixIt(newCity)

}

func canWeFixIt(city models.City) {
	LogToConsole("Yes we can!")
	newBuilding := models.Building{}
	newBuilding.ID, _ = uuid.NewV4()
	newBuilding.BuildDate = settings.LastTime
	newBuilding.Floors = 1
	newBuilding.MaxOccupancy = 20
	newBuilding.Name = "Home"
	newBuilding.Type = models.House
	newBuilding.TopLeft = Point{X: randomdata.Number(city.TopLeft.X, city.BottomRight.X-208), Y: randomdata.Number(city.TopLeft.Y, city.BottomRight.Y-208)}
	newBuilding.BottomRight = Point{X: newBuilding.TopLeft.X + 208, Y: newBuilding.TopLeft.Y + 208}
	newBuilding.CityID = city.ID
	//	err := commonDAO.CreateBuilding(newBuilding)
	//	FailOnError(err, "Failed to create building")
	LogToConsole("Created a new home, Updating City")
	//	err = commonDAO.UpdateCity(city)
	//	FailOnError(err, "Failed to update City")
	justTheTwoOfUs(newBuilding)
}

func justTheTwoOfUs(building models.Building) {
	LogToConsole("You and I")
	male := models.Person{}
	female := models.Person{}
	boyNameGenURL := "http://names.drycodes.com/1?nameOptions=boy_names"
	girlNameGenURL := "http://names.drycodes.com/1?nameOptions=girl_names"
	reqBoy, err := http.NewRequest("GET", boyNameGenURL, nil)
	FailOnError(err, "Error with Boy Name URL")
	reqGirl, err := http.NewRequest("GET", girlNameGenURL, nil)
	FailOnError(err, "Error with Girl Name URL")
	webClient := &http.Client{}
	respBoy, err := webClient.Do(reqBoy)
	FailOnError(err, "Error with Boy Name Request")
	respGirl, err := webClient.Do(reqGirl)
	FailOnError(err, "Error with Girl Name Request")
	defer respBoy.Body.Close()
	defer respGirl.Body.Close()
	if respBoy.StatusCode == http.StatusOK && respGirl.StatusCode == http.StatusOK {
		boybodyBytes, err := ioutil.ReadAll(respBoy.Body)
		FailOnError(err, "Failed to read body")
		mname := string(boybodyBytes)
		mname = strings.TrimPrefix(mname, "[\"")
		mname = strings.TrimSuffix(mname, "\"]")
		male.FirstName = strings.Split(mname, "_")[0]
		male.LastName = strings.Split(mname, "_")[1]
		girlbodyBytes, err := ioutil.ReadAll(respGirl.Body)
		FailOnError(err, "Failed to read body")
		fname := string(girlbodyBytes)
		fname = strings.TrimPrefix(fname, "[\"")
		fname = strings.TrimSuffix(fname, "\"]")
		female.FirstName = strings.Split(fname, "_")[0]
		female.LastName = male.LastName
	}
	male.Birthdate = settings.LastTime.AddDate(-18, 0, 0)
	female.Birthdate = male.Birthdate
	male.ChildrenIDs = []*uuid.UUID{}
	female.ChildrenIDs = male.ChildrenIDs
	male.CurrentBuilding = building.ID
	female.CurrentBuilding = male.CurrentBuilding
	male.CurrentXY = building.TopLeft
	female.CurrentXY = male.CurrentXY
	male.Happiness = 100
	female.Happiness = 100
	male.Health = 100
	female.Health = 100
	male.HomeBuilding = male.CurrentBuilding
	female.HomeBuilding = male.HomeBuilding
	male.ID, _ = uuid.NewV4()
	female.ID, _ = uuid.NewV4()
	male.NewToBuilding = false
	female.NewToBuilding = false
	male.Traveling = false
	female.Traveling = false
	male.WorkBuilding = male.HomeBuilding
	female.WorkBuilding = female.HomeBuilding
	male.Spouse = female.ID
	female.Spouse = male.ID
	//	errM := commonDAO.CreatePerson(male)
	//	FailOnError(errM, "Failed to create male")
	//	errF := commonDAO.CreatePerson(female)
	//	FailOnError(errF, "Failed to create female")

	Logger.Printf("People Names: %s %s, %s %s", male.FirstName, male.LastName, female.FirstName, female.LastName)
}

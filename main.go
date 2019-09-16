package main

import (
	"./Core"
	"./GUI"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"sync"
	"time"
)

func main() {


	// Define the number of cores in our platform
	const numCores = 4

	// Initialize clock from the beginning
	clock := Core.Clock{}
	clock.Init(numCores, time.Second * 9)
	// Create a channel to control the clock with the button
	control := make(chan bool)
	lnext := make(chan bool)
	clock.UserControl = &control
	clock.Next = &lnext
	// Mutex to control gui updates
	mutex := sync.Mutex{}


	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	/// Crea un constructor
	b, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Error:", err)
	}
	// Descargue la ventana del archivo Glade al generador
	err = b.AddFromFile("GUI/SystemMonitor.glade")
	if err != nil {
		log.Fatal("Error:", err)
	}

	// Detect btn start and create its function
	obj, _ := b.GetObject("btnstart")
	start := obj.(*gtk.Button)
	start.Connect("clicked", func() {

		// Create the bus, memory and clock
		memory := Core.Memory{}
		bus := Core.Bus{}

		// Initialize them
		bus.Init()
		memgui := GUI.MemoryGUI{}
		memgui.Init(b)
		memgui.Mutex = &mutex
		memory.Init(&bus, memgui)

		// Create the cores
		cores := [4]Core.Core{}


		// Generate instructions and prepare to start
		// Give gui ptrs to the logic
		gui := GUI.CoreGUI{}
		gui.Init(b, "0")
		gui.Mutex = &mutex
		cores[0].Init(0, &memory, &clock.Clock, gui)
		gui = GUI.CoreGUI{}
		gui.Init(b, "1")
		gui.Mutex = &mutex
		cores[1].Init(1, &memory, &clock.Clock, gui)
		gui = GUI.CoreGUI{}
		gui.Init(b, "2")
		gui.Mutex = &mutex
		cores[2].Init(2, &memory, &clock.Clock, gui)
		gui = GUI.CoreGUI{}
		gui.Init(b, "3")
		gui.Mutex = &mutex
		cores[3].Init(3, &memory, &clock.Clock, gui)


		// Run the cores
		go cores[0].Run()
		go cores[1].Run()
		go cores[2].Run()
		go cores[3].Run()

		// Run memory
		go memory.Run(&clock.MemClock)

		// Run the clock user controlled
		obj, _ := b.GetObject("lblnext")
		lblnext := obj.(*gtk.Label)
		go next(&mutex, lblnext, &lnext)
		go clock.Start(true)

	})

	// Detect btn next and create its function
	obj, _ = b.GetObject("btnnext")
	next := obj.(*gtk.Button)
	next.Connect("clicked", func() {

		control <- true
		println("------------------------------------ Clicked ------------------------------------")

	})


	// Obtener el objeto de la ventana principal por ID
	obj, err = b.GetObject("SystemMonitor")
	if err != nil {
		log.Fatal("Error:", err)
	}

	// Convierta del objeto una ventana de tipo gtk.Window
	// y conéctate a la señal "destruir" para que puedas cerrar
	// aplicación al cerrar la ventana
	win := obj.(*gtk.Window)
	win.Connect("destroy", func() {
		gtk.MainQuit()
		close(control)
	})
	// Mostramos todos los widgets en la ventana.
	win.ShowAll()
	// Ejecutamos el ciclo principal de GTK (para renderizar). Se detendrá cuando
	// gtk.MainQuit () se ejecutará
	gtk.Main()

}

func next(mutex *sync.Mutex, lbl *gtk.Label, next *chan bool){

	for{

		msg := <- *next
		mutex.Lock()
		if msg{
			lbl.SetText("Resuming")
		}else {
			lbl.SetText("Press Next to continue")
		}
		mutex.Unlock()

	}

}
package GUI

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"strconv"
	"sync"
)

type CoreGUI struct {

	Mutex 		*sync.Mutex
	Cache   	[8]*gtk.Label
	Inst    	*gtk.Label
	Miss		*gtk.Label
}

type MemoryGUI struct {
	Mutex 		*sync.Mutex
	Line 	[16]*gtk.Label
}

func (gui *CoreGUI) Init(b *gtk.Builder, id string){

	var obj glib.IObject

	for i := 0; i < 8; i++ {
		// Get the object and save it
		obj,_ = b.GetObject("lblc" + id + strconv.Itoa(i))
		gui.Cache[i] = obj.(*gtk.Label)
	}

	// Assign the instruction label
	obj,_ = b.GetObject("lblinst" + id)
	gui.Inst = obj.(*gtk.Label)

	obj,_ = b.GetObject("lblmiss" + id)
	gui.Miss = obj.(*gtk.Label)

}

func (gui *MemoryGUI) Init(b *gtk.Builder){

	var obj glib.IObject

	for i := 0; i < 16; i++ {
		// Get the object and save it
		obj,_ = b.GetObject("lbll" + strconv.Itoa(i))
		gui.Line[i] = obj.(*gtk.Label)
	}

}


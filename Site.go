package main

import (
	"html/template"
	"net/http"
	"path/filepath"
	"github.com/julienschmidt/httprouter"
	"log"
	"strconv"
)

type WebSite struct {
	app *Application
}

type Action struct {
	Site *WebSite
}
type WelcomeAction struct {
	Action
}
type EchoAction struct {
	Action
}
type ViewAction struct {
	Action
}
type ComposeAction struct {
	Action
}

func (self *WelcomeAction) ServeHTTP(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	lp := filepath.Join("views", "layout.tmpl")
	fp := filepath.Join("views", "welcome.tmpl")
	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		panic(err)
	}
	outParams := make(map[string]interface{})
	outParams["Areas"] = self.Site.app.config.AreaList.Areas
	tmpl.ExecuteTemplate(w, "layout", outParams)
}

func (self *EchoAction) ServeHTTP(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	lp := filepath.Join("views", "layout.tmpl")
	fp := filepath.Join("views", "echo.tmpl")
	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		panic(err)
	}
	//
	echoTag := params.ByName("name")
	log.Printf("echoTag = %v", echoTag)
	//
	area, err1 := self.Site.app.config.AreaList.SearchByName(echoTag)
	if (err1 != nil) {
		panic(err1)
	}
	log.Printf("area = %v", area)
	//
	var msgBase = SquishMessageBase{}
	msgHeaders, err2 := msgBase.ReadBase(area.Path)
	if (err2 != nil) {
		panic(err2)
	}
	//
	outParams := make(map[string]interface{})
	outParams["Areas"] = self.Site.app.config.AreaList.Areas
	outParams["Area"] = area
	outParams["Headers"] = msgHeaders
	tmpl.ExecuteTemplate(w, "layout", outParams)
}

func (self *ViewAction) ServeHTTP(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	lp := filepath.Join("views", "layout.tmpl")
	fp := filepath.Join("views", "view.tmpl")
	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		panic(err)
	}
	//
	echoTag := params.ByName("name")
	log.Printf("echoTag = %v", echoTag)
	//
	area, err1 := self.Site.app.config.AreaList.SearchByName(echoTag)
	if (err1 != nil) {
		panic(err1)
	}
	log.Printf("area = %v", area)
	//
	messageId := params.ByName("msgid")
	var msgId uint64
	msgId, err12 := strconv.ParseUint(messageId, 16, 32)
	log.Printf("err = %v msgid = %d or %x", err12, msgId, msgId)
	//
	var msgBase = SquishMessageBase{}
	msg, err2 := msgBase.ReadMessage(area.Path, uint32(msgId))
	if (err2 != nil) {
		panic(err2)
	}
	//
	outParams := make(map[string]interface{})
	outParams["Areas"] = self.Site.app.config.AreaList.Areas
	outParams["Area"] = area
	outParams["Msg"] = msg
	outParams["Content"] = msg.GetContent()
	tmpl.ExecuteTemplate(w, "layout", outParams)
}

func (self *ComposeAction) ServeHTTP(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
}

func (self *Application) startSite() {
	//
	webSite := new(WebSite)
	webSite.app = self
	//
	router := httprouter.New()
	//
	welcomeAction := new(WelcomeAction)
	welcomeAction.Site = webSite
	router.GET("/", welcomeAction.ServeHTTP)
	//
	echoAction := new(EchoAction)
	echoAction.Site = webSite
	router.GET("/echo/:name", echoAction.ServeHTTP)
	//
	viewAction := new(ViewAction)
	viewAction.Site = webSite
	router.GET("/echo/:name/view/:msgid", viewAction.ServeHTTP)
	//
	composeAction := new(ComposeAction)
	composeAction.Site = webSite
	router.GET("/echo/:name/compose", composeAction.ServeHTTP)
	//
	//fs := http.FileServer(http.Dir("static"))
	//http.Handle("/static/", http.StripPrefix("/static/", fs))
	//
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}

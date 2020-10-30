package tosser

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/vit1251/golden/pkg/charset"
	cmn "github.com/vit1251/golden/pkg/common"
	"github.com/vit1251/golden/pkg/eventbus"
	"github.com/vit1251/golden/pkg/fidotime"
	"github.com/vit1251/golden/pkg/msg"
	"github.com/vit1251/golden/pkg/packet"
	"github.com/vit1251/golden/pkg/registry"
	"github.com/vit1251/golden/pkg/setup"
	"github.com/vit1251/golden/pkg/tmpl"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

type TosserManager struct {
	event          chan bool
	registry      *registry.Container
}

func NewTosserManager(registry *registry.Container) *TosserManager {

	tm := new(TosserManager)
	tm.registry = registry

	tm.checkDirectories()
	//
	tm.event = make(chan bool)

	eventBus := tm.restoreEventBus()
	eventBus.Register(tm)

	return tm
}

func (self *TosserManager) HandleEvent(event string ) {
	log.Printf("Tosser event receive")
	if event == "Toss" {
		self.event <- true
	}
}

func (self *TosserManager) restoreEventBus() *eventbus.EventBus {
	ConfigManagerPtr := self.registry.Get("EventBus")
	if eventBus, ok := ConfigManagerPtr.(*eventbus.EventBus); ok {
		return eventBus
	} else {
		panic("no event bus")
	}
}

func (self *TosserManager) restoreConfigManager() *setup.ConfigManager {
	ConfigManagerPtr := self.registry.Get("ConfigManager")
	if configManager, ok := ConfigManagerPtr.(*setup.ConfigManager); ok {
		return configManager
	} else {
		panic("no config manager")
	}
}

func (self *TosserManager) checkDirectory(cacheSection string) {

	configManager := self.restoreConfigManager()

	cacheDirectory, _ := configManager.Get("main", cacheSection)
	if cacheDirectory == "" {
		log.Printf("Wrong directory: section = %+v", cacheSection)
		storageDirectory := cmn.GetStorageDirectory()
		cacheDirectory = path.Join(storageDirectory, "Fido", cacheSection)
		log.Printf("Construct new directory: section = %+v cacheDirectory = %+v", cacheSection, cacheDirectory)
		configManager.Set("main", cacheSection, cacheDirectory)
	}
	if _, err1 := os.Stat(cacheDirectory); err1 != nil {
		log.Printf("Directory check: name = %v - ERR", cacheSection)
		if os.IsNotExist(err1) {
			log.Printf("Initial create directory: path = %+v", cacheDirectory)
			os.MkdirAll(cacheDirectory, os.ModeDir|0755)
		} else {
			log.Fatal("TosserManager: checkDirectory: err = %+v", err1)
		}
	} else {
		log.Printf("Directory check: name = %v - OK", cacheSection)
	}

}

func (self *TosserManager) checkDirectories() {

	/* Check mailer directory */
	self.checkDirectory("Inbound")
	self.checkDirectory("Outbound")
	self.checkDirectory("TempInbound")
	self.checkDirectory("TempOutbound")
	self.checkDirectory("Temp")

	/* Check FileBox directory */
	self.checkDirectory("FileBox")

}

func (self *TosserManager) Start() {
	go self.run()
}

func (self *TosserManager) processTosser() {
	newTosser := NewTosser(self.registry)
	newTosser.Toss()
}

func (self *TosserManager) run() {
	log.Printf(" * Tosser service start")
	var procIteration int
	for range self.event {
		log.Printf(" * Tosser start (%d)", procIteration)
		self.processTosser()
		log.Printf(" * Tosser complete (%d)", procIteration)
		procIteration += 1
	}
	log.Printf(" * Tosser service stop")
}

func (self *TosserManager) makePacketName() string {
	now := time.Now()
	unixTime := now.Unix()
	log.Printf("unixTime: dec = %d hex = %x", unixTime, unixTime)
	pktName := fmt.Sprintf("%08x.pkt", unixTime)
	log.Printf("pktName: name = %s", pktName)
	return pktName
}

func (self *TosserManager) makeTimeZone() string {
	newTime := time.Now()
	_, offset := newTime.Zone()
	var sign string = "+"
	if offset < 0 {
		offset = -offset
		sign = "-"
	}
	ZHour, Zmin := offset / 3600, offset % 3600
	var newZone string
	if sign == "+" {
		newZone = fmt.Sprintf("%02d%02d", ZHour, Zmin)
	} else if sign == "-" {
		newZone = fmt.Sprintf("-%02d%02d", ZHour, Zmin)
	}
	log.Printf("zone = %s", newZone)
	return newZone
}

/// ORIGIN LENGTH 79 http://ftsc.org/docs/fsc-0068.001
func (self *TosserManager) prepareOrigin(Origin string) string {
	var result string = " -- No origins exist -- "
	/* Check Origin is external */
	if strings.HasPrefix(Origin, "@") {
		newPath := strings.TrimPrefix(Origin, "@")
		content, err := ioutil.ReadFile(newPath)
		if err == nil {
			newContent := string(content)
			rows := strings.Split(newContent, "\n")
			rand.Seed(time.Now().Unix())
			n := rand.Intn(len(rows))
			oneLine := rows[n]
			newOneLine := strings.Trim(oneLine, " \t\n\r")
			result = newOneLine
		}
	} else {
		result = Origin
	}
	if len(result) > 78 {
		result = result[:78]
	}
	return  result
}

func (self *TosserManager) makePacketEchoMessage(em *EchoMessage) (string, error) {

	configManager := self.restoreConfigManager()
	messageManager := self.restoreMessageManager()
	charsetManager := self.restoreCharsetManager()

	/* Create packet name */
	tempOutbound, _ := configManager.Get("main", "TempOutbound")
	password, _ := configManager.Get("main", "Password")

	packetName := self.makePacketName()
	tempPacketPath := path.Join(tempOutbound, packetName)

	/* Open outbound packet */
	pw, err1 := packet.NewPacketWriter(tempPacketPath)
	if err1 != nil {
		return "", err1
	}
	defer pw.Close()

	/* Ask source address */
	myAddr, _ := configManager.Get("main", "Address")
	bossAddr, _ := configManager.Get("main", "Link")
	realName, _ := configManager.Get("main", "RealName")
	TearLine, _ := configManager.Get("main", "TearLine")
	Origin, _ := configManager.Get("main", "Origin")

	/* Write packet header */
	pktHeader := packet.NewPacketHeader()
	pktHeader.OrigAddr.SetAddr(myAddr)
	pktHeader.DestAddr.SetAddr(bossAddr)
	pktHeader.SetPassword(password)

	if err := pw.WritePacketHeader(pktHeader); err != nil {
		return "", err
	}

	/* Encode message headers */
	newSubject, err1 := charsetManager.Encode([]rune(em.Subject))
	if err1 != nil {
		return "", err1
	}
	newTo, err2 := charsetManager.Encode([]rune(em.To))
	if err2 != nil {
		return "", err2
	}
	newFrom, err3 := charsetManager.Encode([]rune(realName))
	if err3 != nil {
		return "", err3
	}

	/* Prepare packet message */
	msgHeader := packet.NewPacketMessageHeader()
	msgHeader.OrigAddr.SetAddr(myAddr)
	msgHeader.DestAddr.SetAddr(bossAddr)
	msgHeader.SetAttribute(packet.PacketAttrDirect)
	msgHeader.SetToUserName(newTo)
	msgHeader.SetFromUserName(newFrom)
	msgHeader.SetSubject(newSubject)
	var now *fidotime.FidoDate = fidotime.NewFidoDate()
	now.SetNow()
	msgHeader.SetTime(now)

	if err := pw.WriteMessageHeader(msgHeader); err != nil {
		return "", err
	}

	/* Message UUID */
	u1 := uuid.NewV4()
	//	u1, err4 := uuid.NewV4()
	//	if err4 != nil {
	//		return err4
	//	}

	/* Prepare origin */
	Origin = self.prepareOrigin(Origin)

	/* Prepare new message */
	t := tmpl.NewTemplate()
	newTearLine, _ := t.Render(TearLine)
	newOrigin, _ := t.Render(Origin)
	newTID, _ := t.Render("Golden/{GOLDEN_PLATFORM} {GOLDEN_VERSION} {GOLDEN_RELEASE_DATE} ({GOLDEN_RELEASE_HASH})")

	/* Construct message content */
	msgContent := messageManager.NewMessageContent(self.registry)

	msgContent.SetCharset("CP866")

	msgContent.AddLine(em.GetBody())
	msgContent.AddLine("")
	msgContent.AddLine(fmt.Sprintf("--- %s", newTearLine))
	msgContent.AddLine(fmt.Sprintf(" * Origin: %s (%s)", newOrigin, myAddr))

	rawMsg := msgContent.Pack()

	/* Calculate checksumm */
	h := crc32.NewIEEE()
	h.Write(rawMsg)
	hs := h.Sum32()
	log.Printf("crc32 = %+v", hs)

	newZone := self.makeTimeZone()

	/* Write message body */
	msgBody := packet.NewMessageBody()
	//
	msgBody.SetArea(em.AreaName)
	//
	msgBody.AddKludge("TZUTC", newZone)
	//msgBody.AddKludge("CHRS", "UTF-8 4")
	msgBody.AddKludge("CHRS", "CP866 2")
	msgBody.AddKludge("MSGID", fmt.Sprintf("%s %08x", myAddr, hs))
	msgBody.AddKludge("UUID", fmt.Sprintf("%s", u1))
	msgBody.AddKludge("TID", newTID)
	if em.Reply != "" {
		msgBody.AddKludge("REPLY", em.Reply)
	}
	//
	msgBody.SetRaw(rawMsg)
	//
	if err5 := pw.WriteMessage(msgBody); err5 != nil {
		return "", err5
	}

	return packetName, nil
}

func (self *TosserManager) WriteEchoMessage(em *EchoMessage) error {

	configManager := self.restoreConfigManager()

	inbound, _ := configManager.Get("main", "Inbound")
	outbound, _ := configManager.Get("main", "Outbound")
	tempOutbound, _ := configManager.Get("main", "TempOutbound")

	packetName, err1 := self.makePacketEchoMessage(em)
	if err1 != nil {
		return err1
	}
	tempPacketPath := path.Join(tempOutbound, packetName)

	// Copy to Inbound and Outbound
	log.Printf("Publish packet: name = %+v", tempPacketPath)
	if err := self.PushPacket(tempPacketPath, path.Join(inbound, packetName)); err != nil {
		log.Printf("Fail on copy Inbound: err = %+v", err)
	}
	if err := self.PushPacket(tempPacketPath, path.Join(outbound, packetName)); err != nil {
		log.Printf("Fail on copy Outbound: err = %+v", err)
	}

	return nil
}

func (self *TosserManager) PushPacket(src string, dst string) error {

	log.Printf("Publish packet: %s -> %s", src, dst)

	source, err1 := os.Open(src)
	if err1 != nil {
		return err1
	}
	defer source.Close()

	destination, err2 := os.Create(dst)
	if err2 != nil {
		return err2
	}
	defer destination.Close()

	nBytes, err3 := io.Copy(destination, source)
	log.Printf("Copy %+v", nBytes)

	return err3
}

func (self *TosserManager) WriteNetmailMessage(nm *NetmailMessage) error {

	configManager := self.restoreConfigManager()
	messageManager := self.restoreMessageManager()
	charsetManager := self.restoreCharsetManager()

	var params struct {
		Outbound string
		From string
		FromName string
		TearLine string
		Origin string
	}

	/* Create packet name */
	params.Outbound, _ = configManager.Get("main", "Outbound")
	params.From, _ = configManager.Get("main", "Address")
	params.FromName, _ = configManager.Get("main", "RealName")

	origin, _ := configManager.Get("main", "Origin")
	origin1 := self.prepareOrigin(origin)
	params.Origin = origin1

	TearLine, _ := configManager.Get("main", "TearLine")
	params.TearLine = TearLine

	/* Create packet name */
	pktName := self.makePacketName()
	name := path.Join(params.Outbound, pktName)
	log.Printf("Write Netmail packet %s", name)

	/* Open outbound packet */
	pw, err1 := packet.NewPacketWriter(name)
	if err1 != nil {
		return err1
	}
	defer pw.Close()

	/* Write packet header */
	pktHeader := packet.NewPacketHeader()
	pktHeader.OrigAddr.SetAddr(params.From)
	pktHeader.DestAddr.SetAddr(nm.ToAddr)

	if err := pw.WritePacketHeader(pktHeader); err != nil {
		return err
	}

	/* Encode message */
	newSubject, err1 := charsetManager.Encode([]rune(nm.Subject))
	if err1 != nil {
		return err1
	}
	newTo, err2 := charsetManager.Encode([]rune(nm.To))
	if err2 != nil {
		return err2
	}
	newFrom, err3 := charsetManager.Encode([]rune(params.FromName))
	if err3 != nil {
		return err3
	}

	/* Prepare packet message */
	msgHeader := packet.NewPacketMessageHeader()
	msgHeader.OrigAddr.SetAddr(params.From)
	msgHeader.DestAddr.SetAddr(nm.ToAddr)
	msgHeader.SetAttribute(packet.PacketAttrDirect)
	msgHeader.SetToUserName(newTo)
	msgHeader.SetFromUserName(newFrom)
	msgHeader.SetSubject(newSubject)
	var now *fidotime.FidoDate = fidotime.NewFidoDate()
	now.SetNow()
	msgHeader.SetTime(now)

	if err := pw.WriteMessageHeader(msgHeader); err != nil {
		return err
	}

	/* Message UUID */
	u1 := uuid.NewV4()
	//	u1, err4 := uuid.NewV4()
	//	if err4 != nil {
	//		return err4
	//	}

	/* Prepare new message */
	t := tmpl.NewTemplate()
	newTearLine, _ := t.Render(params.TearLine)
	newOrigin, _ := t.Render(params.Origin)
	newTID, _ := t.Render("Golden/{GOLDEN_PLATFORM} {GOLDEN_VERSION} {GOLDEN_RELEASE_DATE} ({GOLDEN_RELEASE_HASH})")

	/* Construct message content */
	msgContent := messageManager.NewMessageContent(self.registry)
	msgContent.SetCharset("CP866")
	msgContent.AddLine(nm.GetBody())
	msgContent.AddLine("")
	msgContent.AddLine(fmt.Sprintf("--- %s", newTearLine))
	msgContent.AddLine(fmt.Sprintf(" * Origin: %s (%s)", newOrigin, params.From))
	rawMsg := msgContent.Pack()

	/* Calculate checksumm */
	h := crc32.NewIEEE()
	h.Write(rawMsg)
	hs := h.Sum32()
	log.Printf("crc32 = %+v", hs)

	/* Write message body */
	msgBody := packet.NewMessageBody()

	/* Cross network NETMAIL */
	destinationAddress := fmt.Sprintf("%d:%d/%d", msgHeader.DestAddr.Zone, msgHeader.DestAddr.Net, msgHeader.DestAddr.Node)
	originAddress := fmt.Sprintf("%d:%d/%d", msgHeader.OrigAddr.Zone, msgHeader.OrigAddr.Net,  msgHeader.OrigAddr.Node)

	msgBody.AddKludge("INTL", fmt.Sprintf("%s %s", destinationAddress, originAddress))
	msgBody.AddKludge("FMPT", fmt.Sprintf("%d", msgHeader.OrigAddr.Point))
	msgBody.AddKludge("TOPT", fmt.Sprintf("%d", msgHeader.DestAddr.Point))
	msgBody.AddKludge("CHRS", "CP866 2")
	msgBody.AddKludge("MSGID", fmt.Sprintf("%s %08x", params.From, hs))
	msgBody.AddKludge("UUID", fmt.Sprintf("%s", u1))
	msgBody.AddKludge("TID", newTID)

	/* Set message body */
	msgBody.SetRaw(rawMsg)

	/* Write message in packet */
	if err := pw.WriteMessage(msgBody); err != nil {
		return err
	}

	return nil
}

func (self *TosserManager) restoreMessageManager() *msg.MessageManager {
	return nil
}

func (self *TosserManager) restoreCharsetManager() *charset.CharsetManager {
	return nil
}

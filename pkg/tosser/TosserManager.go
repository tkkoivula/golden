package tosser

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/vit1251/golden/pkg/msg"
	"github.com/vit1251/golden/pkg/packet"
	"github.com/vit1251/golden/pkg/setup"
	"hash/crc32"
	"log"
	"path"
	"time"
)

type TosserManager struct {
	SetupManager *setup.SetupManager
}

type NetmailMessage struct {
	Subject string
	To      string
	ToAddr  string
	From    string
	Body    string
}

type EchoMessage struct {
	Subject string
	To string
	From string
	Body string
	AreaName string
}

func (m *EchoMessage) SetSubject(subj string) {
	m.Subject = subj
}

func NewTosserManager(sm *setup.SetupManager) *TosserManager {
	tm := new(TosserManager)
	tm.SetupManager = sm
	return tm
}

func (self *TosserManager) makePacketName() string {
	/* Message UUID */
	u1 := uuid.NewV4()
	//	u1, err4 := uuid.NewV4()
	//	if err4 != nil {
	//		return err4
	//	}

	/* Create packet name */
	pktName := fmt.Sprintf("%s.pkt", u1)

	return pktName
}

func (self *TosserManager) WriteEchoMessage(em *EchoMessage) error {

	/* Create packet name */
	outb, err1 := self.SetupManager.Get("main", "Outbound", ".")
	if err1 != nil {
		return err1
	}

	pktName := self.makePacketName()
	name := path.Join(outb, pktName)

	/* Open outbound packet */
	pw, err1 := packet.NewPacketWriter(name)
	if err1 != nil {
		return err1
	}
	defer pw.Close()

	/* Ask source address */
	myAddr, err2 := self.SetupManager.Get("main", "Address", "0:0/0.0")
	if err2 != nil {
		return err2
	}
	bossAddr, err3 := self.SetupManager.Get("main", "Link", "0:0/0.0")
	if err3 != nil {
		return err3
	}
	realName, err4 := self.SetupManager.Get("main", "RealName", "John Smith")
	if err4 != nil {
		return err4
	}
	TearLine, err5 := self.SetupManager.Get("main", "TearLine", "John Smith")
	if err5 != nil {
		return err5
	}
	Origin, err6 := self.SetupManager.Get("main", "Origin", "John Smith")
	if err6 != nil {
		return err6
	}

	/* Write packet header */
	pktHeader := packet.NewPacketHeader()
	pktHeader.OrigAddr.SetAddr(myAddr)
	pktHeader.DestAddr.SetAddr(bossAddr)

	if err := pw.WritePacketHeader(pktHeader); err != nil {
		return err
	}

	/* Prepare packet message */
	msgHeader := packet.NewPacketMessageHeader()
	msgHeader.OrigAddr.SetAddr(myAddr)
	msgHeader.DestAddr.SetAddr(bossAddr)
	msgHeader.SetAttribute("Direct")
	msgHeader.SetToUserName(em.To)
	msgHeader.SetFromUserName(realName)
	msgHeader.SetSubject(em.Subject)
	var now time.Time = time.Now()
	msgHeader.SetTime(&now)

	if err := pw.WriteMessageHeader(msgHeader); err != nil {
		return err
	}

	/* Message UUID */
	u1 := uuid.NewV4()
	//	u1, err4 := uuid.NewV4()
	//	if err4 != nil {
	//		return err4
	//	}

	/* Construct message content */
	msgContent := msg.NewMessageContent()

	msgContent.SetCharset("CP866")

	msgContent.AddLine(em.Body)
	msgContent.AddLine("")
	msgContent.AddLine(fmt.Sprintf("--- %s", TearLine))
	msgContent.AddLine(fmt.Sprintf(" * Origin: %s (%s)", Origin, myAddr))

	rawMsg := msgContent.Pack()

	/* Calculate checksumm */
	h := crc32.NewIEEE()
	h.Write(rawMsg)
	hs := h.Sum32()
	log.Printf("crc32 = %+v", hs)

	/* Write message body */
	msgBody := packet.NewMessageBody()
	//
	msgBody.SetArea(em.AreaName)
	//
	//msgBody.AddKludge("TZUTC", "0300")
	//msgBody.AddKludge("CHRS", "UTF-8 4")
	msgBody.AddKludge("CHRS", "CP866 2")
	msgBody.AddKludge("MSGID", fmt.Sprintf("%s %08x", myAddr, hs))
	msgBody.AddKludge("UUID", fmt.Sprintf("%s", u1))
	msgBody.AddKludge("TID", "golden/win 1.2.8 2020-02-18 13:19 MSK (master)")
	//
	msgBody.SetRaw(rawMsg)
	//
	if err5 := pw.WriteMessage(msgBody); err5 != nil {
		return err5
	}

	return nil
}

func (self *TosserManager) WriteNetmailMessage(nm *NetmailMessage) error {

	var params struct {
		Outbound string
		From string
		FromName string
		TearLine string
		Origin string
	}

	/* Create packet name */
	if outb, err := self.SetupManager.Get("main", "Outbound", "."); err == nil {
		params.Outbound = outb
	} else {
		return err
	}
	if from, err := self.SetupManager.Get("main", "Address", "."); err == nil {
		params.From = from
	} else {
		return err
	}
	if fromName, err := self.SetupManager.Get("main", "RealName", "John Smith"); err == nil {
		params.FromName = fromName
	} else {
		return err
	}
	if origin, err := self.SetupManager.Get("main", "Origin", "Empty"); err == nil {
		params.Origin = origin
	} else {
		return err
	}
	if TearLine, err := self.SetupManager.Get("main", "TearLine", "Empty"); err == nil {
		params.TearLine = TearLine
	} else {
		return err
	}

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

	/* Prepare packet message */
	msgHeader := packet.NewPacketMessageHeader()
	msgHeader.OrigAddr.SetAddr(params.From)
	msgHeader.DestAddr.SetAddr(nm.ToAddr)
	msgHeader.SetAttribute("Direct")
	msgHeader.SetToUserName(nm.To)
	msgHeader.SetFromUserName(params.FromName)
	msgHeader.SetSubject(nm.Subject)
	var now time.Time = time.Now()
	msgHeader.SetTime(&now)

	if err := pw.WriteMessageHeader(msgHeader); err != nil {
		return err
	}

	/* Message UUID */
	u1 := uuid.NewV4()
	//	u1, err4 := uuid.NewV4()
	//	if err4 != nil {
	//		return err4
	//	}

	/* Construct message content */
	msgContent := msg.NewMessageContent()
	msgContent.SetCharset("CP866")
	msgContent.AddLine(nm.Body)
	msgContent.AddLine("")
	msgContent.AddLine(fmt.Sprintf("--- %s", params.TearLine))
	msgContent.AddLine(fmt.Sprintf(" * Origin: %s (%s)", params.Origin, params.From))
	rawMsg := msgContent.Pack()

	/* Calculate checksumm */
	h := crc32.NewIEEE()
	h.Write(rawMsg)
	hs := h.Sum32()
	log.Printf("crc32 = %+v", hs)

	/* Write message body */
	msgBody := packet.NewMessageBody()

	//
	//msgBody.AddKludge("TZUTC", "0300")
	msgBody.AddKludge("CHRS", "CP866 2")
	msgBody.AddKludge("MSGID", fmt.Sprintf("%s %08x", params.From, hs))
	msgBody.AddKludge("UUID", fmt.Sprintf("%s", u1))
	msgBody.AddKludge("TID", "golden/win 1.2.8 2020-02-18 13:19 MSK (master)")
	msgBody.AddKludge("FMPT", fmt.Sprintf("%d", msgHeader.OrigAddr.Point))
	msgBody.AddKludge("TOPT", fmt.Sprintf("%d", msgHeader.DestAddr.Point))

	/* Set message body */
	msgBody.SetRaw(rawMsg)

	/* Write message in packet */
	if err := pw.WriteMessage(msgBody); err != nil {
		return err
	}

	return nil
}

func (self *TosserManager) NewEchoMessage() *EchoMessage {
	em := new(EchoMessage)
	return em
}

func (self *TosserManager) NewNetmailMessage() *NetmailMessage {
	nm := new(NetmailMessage)
	return nm
}
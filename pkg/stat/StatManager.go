package stat

import (
	"github.com/vit1251/golden/pkg/setup"
	"log"
)

type StatManager struct {
	Path string
}

type Stat struct {
	TicReceived      int
	TicSent          int
	EchomailReceived int
	EchomailSent     int
	NetmailReceived  int
	NetmailSent      int

	Dupe int

	PacketReceived  int
	PacketSent      int

	MessageReceived int
	MessageSent     int

}

var stat Stat

func NewStatManager() (*StatManager) {
	sm := new(StatManager)
	sm.Path = setup.GetBasePath()
	return sm
}

func (self *StatManager) RegisterNetmail(filename string) (error) {
	log.Printf("Stat: RegisterNetmail: %s", filename)
	stat.NetmailReceived += 1
	return nil
}

func (self *StatManager) RegisterARCmail(filename string) (error) {
	log.Printf("Stat: RegisterARCmail: %s", filename)
	stat.EchomailReceived += 1
	return nil
}

func (self *StatManager) RegisterFile(filename string) (error) {
	log.Printf("Stat: RegisterFile: %s", filename)
	return nil
}

func (self *StatManager) GetStat() (*Stat, error) {
	return &stat, nil
}

func (self *StatManager) RegisterPacket(p string) error {
	stat.PacketReceived += 1
	return nil
}

func (self *StatManager) RegisterDupe() error {
	stat.Dupe += 1
	return nil
}

func (self *StatManager) RegisterMessage() error {
	stat.MessageReceived += 1
	return nil
}
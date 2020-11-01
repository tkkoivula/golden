package echomail

import "strings"

type Area struct {
	name            string     /* Echo name              */
	Summary         string     /* Echo summary           */
	Charset         string     /* Echo charset           */
	Flag            string     /* Echo marker            */
	Path            string     /* Echo directory         */
	MessageCount    int        /* Echo message count     */
	NewMessageCount int        /* Echo new message count */
}

func NewArea() *Area {
	a := new(Area)
	a.Charset = "CP866"
	a.Flag = "A"
	return a
}

func (self *Area) Name() string {
	return self.name
}

func (self *Area) SetName(name string) {
	self.name = strings.ToUpper(name)
}
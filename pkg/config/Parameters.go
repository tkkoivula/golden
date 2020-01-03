package config

type ParamType int

const ParamString ParamType = 1

type Param struct {
	Summary    string         /* Parameter summary     */
	Section    string         /* Parameter section     */
	Name       string         /* Parameter name        */
	Value      string         /* Parameter value       */
	IsSet      bool           /* Parameter exists mark */
	Type       ParamType      /* Parameter value type  */
}

type ParamStorage struct {
	Path     string  /* Param stroage path */
	Params []Param   /* Param array        */
}

func NewParamStorage() (*ParamStorage, error) {

	ps := new(ParamStorage)
	ps.Path = "~/.golden.sqlite3"

	/* Initialize parameters */
	ps.Register("Name", "")
	ps.Register("Sysop", "")
	ps.Register("Location", "")
	ps.Register("Address", "")
	ps.Register("Inbound", "")
	ps.Register("ProtInbound", "")
	ps.Register("Outbound", "")
	ps.Register("Origin", "Add this Origin to hpt messages: post, reports about new areas created.")
	ps.Register("SortEchoList", "This keyword determines sorting mode for echolist. By default echolist is sorted by name.")
	ps.Register("TearLine", "Add this tearline to messages (post, reports about new areas created)")
	ps.Register("MsgBaseDir", "This command specifies the path where msgBases of autocreated areas are stored.")
	ps.Register("TempInbound", "This command specifies a path which is used while tossing. The incoming packets are unpacked there.")
	ps.Register("TempOutbound", "This command specifies your temporary outbound path. It is used for storing outgoing pkt-files before packing.")

	/* Done */
	return ps, nil
}

func (self *ParamStorage) Set(name string, value string) (error) {
	return nil
}

func (self *ParamStorage) Get(name string, defaultValue string) (value string, error) {
	return "", nil
}

func (self *ParamStorage) Register(name string, summary string) (error) {
	return nil
}

func (self *ParamStorage) Audit(msg string) (error) {

	/* Store audit message in parameter storage */

	return nil

}

func (self *ParamStorage) Restore() (error) {

	/* Open SQL storage */
	db, err1 := sql.Open("sqlite3", self.Path)
		return nil, err
	}
	defer db.Close()

	/* Restore parameters */

	return nil
}

func (self *ParamStorage) Store() (error) {

	/* Open SQL storage */
	db, err1 := sql.Open("sqlite3", self.Path)
		return nil, err
	}
	defer db.Close()

	/* Store parameters */

	return nil

}

package gqrx

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	GQRX_Get_Freq         = "f"
	GQRX_Set_Freq         = "F"
	GQRX_Get_Mod          = "m"
	GQRX_Set_Mod          = "M"
	GQRX_Get_Sig_Strength = "l STRENGTH"
	GQRX_Get_Sql          = "l SQL"
	GQRX_Set_Sql          = "L SQL"
	GQRX_Close_conn       = "Q"
	GQRX_Get_Mute         = "u MUTE"
	GQRX_Set_Mute         = "U MUTE"
)

var (
	Demods = []string{"OFF", "RAW", "AM", "AMS", "LSB", "USB", "CWL", "CWR", "CWU", "CW", "FM", "WFM", "WFM_ST", "WFM_ST_OIRT"}
)

type client struct {
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	addr    string
	msgChan chan string
}

// NewClient : returns  client
func NewClient(addr string) *client {
	newClient := &client{addr: addr}
	newClient.msgChan = make(chan string)
	return newClient
}

func (c *client) Connect() error {
	var err error
	c.conn, err = net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	c.reader = bufio.NewReader(c.conn)
	c.writer = bufio.NewWriter(c.conn)
	go c.listen()
	return nil
}

func (c *client) Disconnect() {
	// TODO : stop listener
	// TODO : clear channel
	// TODO : disconnect
	fmt.Println(c.getString(GQRX_Close_conn))
}

func (c *client) SetDemod(mode string, bandwidth int64) error {
	demodExists := false
	for _, demod := range Demods {
		if demod == mode {
			demodExists = true
		}
	}
	if demodExists == false {
		return fmt.Errorf("Mode not found: %s\n  Avaliable modes:\n   %v", mode, Demods)
	}
	_, err := c.getString(GQRX_Set_Mod + fmt.Sprintf(" %s %d", mode, bandwidth))
	if err != nil {
		return fmt.Errorf("failed to set mode error: %v", err)
	}
	return nil
}

// GetDspStatus Return the status if the DSP
func (c *client) GetDspStatus() (bool, error) {
	dspStatus, err := c.getInt64("u DSP")
	if err != nil {
		return false, err
	}
	if dspStatus == 1 {
		return true, nil
	}
	return false, nil
}

// SetDspStatus true = play, false = stop
func (c *client) SetDspStatus(status bool) error {
	playCommand := "U DSP 0"
	if status {
		playCommand = "U DSP 1"
	}
	if _, err := c.getString(playCommand); err != nil {
		return err
	}
	return nil
}

// GetMute
func (c *client) GetMute() (bool, error) {
	muteStatus, err := c.getInt64(GQRX_Get_Mute)
	if err != nil {
		return false, fmt.Errorf("failed to get mute status: %v", err)
	}
	if muteStatus == 0 {
		return false, nil
	}
	return true, nil
}

// SetMute
func (c *client) SetMute(input bool) error {
	muteStatus := 0
	if input {
		muteStatus = 1
	}
	if _, err := c.getString(fmt.Sprintf("%s %d", GQRX_Set_Mute, muteStatus)); err != nil {
		return fmt.Errorf("failed to set mute status: %v", err)
	}
	return nil
}

// GetSql
func (c *client) GetSql() (float64, error) {
	return c.getFloat(GQRX_Get_Sql)
}

// SetSql
func (c *client) SetSql(input float64) error {
	if _, err := c.getString(fmt.Sprintf("%s %.2f", GQRX_Set_Sql, input)); err != nil {
		return fmt.Errorf("failed to set sql: %v", err)
	}
	return nil
}

// GetSigStrength yep!
func (c *client) GetSigStrength() (float64, error) {
	return c.getFloat(GQRX_Get_Sig_Strength)
}

// SetFreq wonder What this does?
func (c *client) SetFreq(freq int64) error {
	_, err := c.getString(GQRX_Set_Freq + fmt.Sprintf(" %d", freq))
	if err != nil {
		return err
	}
	return nil
}

// GetFreq return
func (c *client) GetFreq() (int64, error) {
	return c.getInt64(GQRX_Get_Freq)
}

////////////////////~~dev funcs~~///////////////////////////

// GetOpts arg (u, l)
func (c *client) GetOpts(input string) (string, error) {
	return c.getString(fmt.Sprintf("%s ?", input))
}

// GetTestValue string
func (c *client) GetTestValue(input string) (string, error) {
	return c.getString(fmt.Sprintf("%s", input))
}

////////////////////~~Private funcs~~///////////////////////////

func (c *client) getFloat(msg string) (float64, error) {
	if err := c.sendMsg(msg); err != nil {
		return 0.0, err
	}
	value := <-c.msgChan
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("Error converting value to int: %v\n", err)

	}
	return f, nil
}

func (c *client) getInt64(msg string) (int64, error) {
	if err := c.sendMsg(msg); err != nil {
		return 0, err
	}
	value := <-c.msgChan
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Error converting value to int: %v\n", err)
	}
	return i, nil
}

func (c *client) getString(msg string) (string, error) {
	if err := c.sendMsg(msg); err != nil {
		return "", err
	}
	value := <-c.msgChan
	return value, nil
}

func (c *client) sendMsg(msg string) (err error) {
	_, err = c.writer.WriteString(fmt.Sprintf("%s\r\n", msg))
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}
	c.writer.Flush()
	return
}

func (c *client) listen() {
	go func() {
		for {
			res, err := c.reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading c.msgChan:", err)
				return
			}
			c.msgChan <- strings.Replace(res, "\n", "", 1)
		}
	}()
}

package gqrx

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	GQRX_Get_Freq         = "f"
	GQRX_Set_Freq         = "F"
	GQRX_Get_Demod        = "m"
	GQRX_Set_Demod        = "M"
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

type Client struct {
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	addr    string
	msgChan chan string
}

// NewClient
// returns  Client
// Default GQRX port : 7356
//
//	   addr := "127.0.0.1:7356"
//	   client := gqrx.NewClient(addr)
//	   if err := client.Connect(); err != nil {
//		      log.Fatalf("Error connecting: %v", err)
//	   }
func NewClient(addr string) *Client {
	newClient := &Client{addr: addr}
	newClient.msgChan = make(chan string)
	return newClient
}

// Connect
//
//	 Connect to GQRX
//
//	   if err := client.Connect(); err != nil {
//		      log.Fatalf("Error connecting: %v", err)
//	   }
func (c *Client) Connect() error {
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

// Disconnect
// Send disconnect command to GQRX
// close net connection
//
//	if err := client.Disconnect(); err != nil {
//	    log.Fatalf("Error disconnecting: %v", err)
//	}
func (c *Client) Disconnect() error {
	if err := c.sendMsg(GQRX_Close_conn); err != nil {
		return fmt.Errorf("gqrx: disconnect failed: %v", err)
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("c.conn.Close(): disconnect failed: %v", err)
	}
	return nil
}

// SetDemod (mode string, bandwidth int64)
//
//	if err := client.SetDemod("WFM_ST", 160000); err != nil {
//		return
//	}
func (c *Client) SetDemod(mode string, bandwidth int64) error {
	demodExists := false
	for _, demod := range Demods {
		if demod == mode {
			demodExists = true
		}
	}
	if demodExists == false {
		return fmt.Errorf("Mode not found: %s\nAvaliable modes:\n	%v", mode, Demods)
	}
	_, err := c.getString(GQRX_Set_Demod + fmt.Sprintf(" %s %d", mode, bandwidth))
	if err != nil {
		return fmt.Errorf("failed to set mode error: %v", err)
	}
	return nil
}

// GetDemod
// return mode as string and mode bandwidth as int64
//
//	mode, bandwidth, err := client.GetDemod()
//	if err != nil {
//	    log.Fatalf("Error getting current mode: %v", err)
//	}
//	fmt.Printf("Mode: %s\nBandwidth: %d\n", mode, bandwidth)
func (c *Client) GetDemod() (string, int64, error) {
	// Gqrx returns two lines, line 1 = mode string, line 2 = mode bandwidth
	modValue, err := c.getString(GQRX_Get_Demod)
	if err != nil {
		return "", 0, err
	}
	bandwidthInt, err := strconv.ParseInt(<-c.msgChan, 10, 64)
	if err != nil {
		return "", 0, err
	}
	return modValue, bandwidthInt, nil
}

// GetDspStatus
// Return the status if the DSP
//
//	dspStatus, err := client.GetDspStatus()
//	if err != nil {
//	   log.Fatalf("Error getting DSP status: %v", err)
//	}
//	if dspStatus {
//	    fmt.Println("DSP is running")
//	}
func (c *Client) GetDspStatus() (bool, error) {
	dspStatus, err := c.getInt64("u DSP")
	if err != nil {
		return false, err
	}
	if dspStatus == 1 {
		return true, nil
	}
	return false, nil
}

// SetDspStatus
// true = start
// false = stop
//
//	   // turn on DSP
//	   err := client.SetDspStatus(true)
//	   if err != nil {
//		      log.Fatalf("Error setting DSP status: %v", err)
//	   }
func (c *Client) SetDspStatus(status bool) error {
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
// Get DSP mute status
// true = running, false = stopped
//
//		muted, err := client.GetMute()
//		if err != nil {
//		    log.Fatalf("failed to get mute status: %v", err)
//		}
//		if muted {
//	        fmt.Println("DSP is muted")
//		}
func (c *Client) GetMute() (bool, error) {
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
// true = mute, false = unmute
//
//	if err := client.SetMute(false); err != nil {
//	    log.Fatalf("Error setting mute: %v", err)
//	}
func (c *Client) SetMute(input bool) error {
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
// returns SQL as float64
//
//	sql, err := client.GetSql()
//	if err != nil {
//	    log.Fatalf("Error getting sql: %v", err)
//	}
//	fmt.Printf("SQL: %.2f\n", sql)
func (c *Client) GetSql() (float64, error) {
	return c.getFloat(GQRX_Get_Sql)
}

// SetSql
// set SQL float64
//
//	if err := client.SetSql(64.0); err != nil {
//	    log.Fatalf("Error setting sql: %v", err)
//	}
func (c *Client) SetSql(input float64) error {
	if _, err := c.getString(fmt.Sprintf("%s %.2f", GQRX_Set_Sql, input)); err != nil {
		return fmt.Errorf("failed to set sql: %v", err)
	}
	return nil
}

// GetSigStrength
// Return signal strength as float64
//
//	strength, err := client.GetSigStrength()
//	if err != nil {
//	    return
//	}
//	fmt.Printf("Strength: %.2f\n", strength)
func (c *Client) GetSigStrength() (float64, error) {
	return c.getFloat(GQRX_Get_Sig_Strength)
}

// SetFreq
//
//	if err := client.SetFreq(station); err != nil {
//	    log.Fatalf("Error setting Freq: %v", err)
//	}
func (c *Client) SetFreq(freq int64) error {
	_, err := c.getString(GQRX_Set_Freq + fmt.Sprintf(" %d", freq))
	if err != nil {
		return err
	}
	return nil
}

// GetFreq
//
//	freq, err := client.GetFreq()
//	if err != nil {
//	    log.Fatalf("Error getting frequency: %v", err)
//	}
//	fmt.Printf("Freq: %d\n", freq)
func (c *Client) GetFreq() (int64, error) {
	return c.getInt64(GQRX_Get_Freq)
}

////////////////////~~dev funcs~~///////////////////////////

// GetTestValue string
func (c *Client) GetUserValue(input string) (string, error) {
	return c.getString(input)
}

////////////////////~~Private funcs~~///////////////////////////

func (c *Client) getFloat(msg string) (float64, error) {
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

func (c *Client) getInt64(msg string) (int64, error) {
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

func (c *Client) getString(msg string) (string, error) {
	if err := c.sendMsg(msg); err != nil {
		return "", err
	}
	value := <-c.msgChan
	return value, nil
}

func (c *Client) sendMsg(msg string) (err error) {
	_, err = c.writer.WriteString(fmt.Sprintf("%s\r\n", msg))
	if err != nil {
		return fmt.Errorf("error sending command: %v", err)

	}
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return
}

func (c *Client) listen() {
	go func() {
		for {
			res, err := c.reader.ReadString('\n')
			if err != nil {
				log.Printf("Error reading c.msgChan: %v", err)
				return
			}
			c.msgChan <- strings.Replace(res, "\n", "", 1)
		}
	}()
}

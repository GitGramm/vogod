package optolink

import (
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io"
	"strconv"
	"strings"
)

// xDataPointType is a type to hold raw information from xml unmarshalling
type xDataPointType struct {
	ID                          string `xml:"ID"`
	EventtTypeList              string `xml:"EventTypeList"`
	Description                 string `xml:"Description"`
	Identification              string `xml:"Identification"`
	IdentificationExtension     string `xml:"IdentificationExtension"`
	IdentificationExtensionTill string `xml:"IdentificationExtensionTill"`
}

// EventType holds low-level info for commands like address, data format and conversion hints
type xEventType struct {
	ID          string
	Address     string
	Description string

	FCRead  string
	FCWrite string

	Parameter    string
	SDKDataType  string
	PrefixRead   string
	PrefixWrite  string
	BlockLength  string
	BlockFactor  string
	MappingType  string
	BytePosition string
	ByteLength   string
	BitPosition  string
	BitLength    string

	ALZ string // AuslieferZuStand

	Conversion string

	ConversionFactor string
	ConversionOffset string
	LowerBorder      string
	UpperBorder      string
	Stepping         string // TODO: check if this is given implicitely by conversion

	ValueList string
	Unit      string
}

// ErrNotFound is returned when data could not be found
var ErrNotFound = errors.New("Not found")

// FindDataPointType reads DataPoint info from xml in a format similar to VitoSofts ecnDataPointType.xml format
func FindDataPointType(xmlReader io.Reader, sysDeviceIdent [8]byte, dpt *DataPointType) error {
	decoder := xml.NewDecoder(xmlReader)
	var dp xDataPointType

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "DataPointType" {
				var d xDataPointType
				decoder.DecodeElement(&d, &se)

				// TODO: Should we return any matching device regardless of wether it can be handled via KW or P300?
				if len(d.Identification) != 4 {
					break
				}

				if (len(d.IdentificationExtension) == 0 || (len(d.IdentificationExtension) >= 4 && len(d.IdentificationExtension) <= 6)) &&
					(len(d.IdentificationExtensionTill) == 0 || (len(d.IdentificationExtensionTill) >= 4 && len(d.IdentificationExtensionTill) <= 6)) {

					i, err := strconv.ParseUint("0x00"+d.Identification, 0, 16)
					if err != nil {
						break
					}

					// Match sysDeviceGroupIdent
					if sysDeviceIdent[0] != byte((i>>8)&0xff) {
						break
					}
					// Match sysDeviceIdent
					if sysDeviceIdent[1] != byte(i&0xff) {
						break
					}

					idExt, err := strconv.ParseUint("0x00"+d.IdentificationExtension, 0, 24)
					if err != nil {
						idExt = 0

					}
					idExtTill, err := strconv.ParseUint("0x00"+d.IdentificationExtensionTill, 0, 24)
					if err != nil {
						idExtTill = 0
					}

					var dataPointIDExt uint64
					dataPointIDExt = uint64(sysDeviceIdent[2])<<8 | uint64(sysDeviceIdent[3])
					if (len(d.IdentificationExtension) > 4) || (len(d.IdentificationExtensionTill) > 4) {
						dataPointIDExt = uint64(dataPointIDExt)<<16 | uint64(sysDeviceIdent[4])<<8 | uint64(sysDeviceIdent[5])
					}
					if dataPointIDExt >= idExt && (dataPointIDExt < idExtTill || idExtTill == 0) {
						if dp.ID == "" {
							// First match, break as there is nothing to compare
							dp = d
							break
						}

						dpidExt, err := strconv.ParseUint("0x00"+dp.IdentificationExtension, 0, 24)
						if err != nil {
							dpidExt = 0
						}
						dpidExtTill, err := strconv.ParseUint("0x00"+dp.IdentificationExtensionTill, 0, 24)
						if err != nil {
							dpidExtTill = 0
						}

						if idExt >= dpidExt && (idExtTill < dpidExtTill || dpidExtTill == 0) {
							// A better match than the previously found one
							dp = d
						}
					}
				}
			}
		default:
			//
		}
	}
	if dp.ID != "" {
		r := dpt
		r.ID = dp.ID
		r.Description = dp.Description
		r.SysDeviceIdent = sysDeviceIdent
		etl := dpt.EventTypes

		for _, et := range strings.Split(dp.EventtTypeList, ";") {
			etl[et] = EventType{ID: et}
		}
		r.EventTypes = etl

		return nil
	}
	return ErrNotFound
}

// FindEventTypes reads EventType infos from xml in a format similar to VitoSofts ecnEventType.xml format
func FindEventTypes(xmlReader io.Reader, etl *EventTypeList) int {
	decoder := xml.NewDecoder(xmlReader)
	found := 0

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "EventType" {
				var et xEventType
				decoder.DecodeElement(&et, &se)

				if _, ok := (*etl)[et.ID]; !ok {
					break
				}
				(*etl)[et.ID], _ = validatexEventType(et)
				found++

			}
		default:
			//
		}
	}
	return found
}

func validatexEventType(xet xEventType) (EventType, error) {
	var et EventType

	et.ID = xet.ID
	i, err := strconv.ParseUint(xet.Address, 0, 16)
	if err == nil {
		et.Address = uint16(i)
	}
	et.Description = xet.Description
	et.FCRead = str2CmdType(xet.FCRead)
	et.FCWrite = str2CmdType(xet.FCWrite)

	et.Parameter = xet.Parameter
	et.SDKDataType = xet.SDKDataType

	p, err := hex.DecodeString(xet.PrefixRead)
	if err == nil {
		et.PrefixRead = p
	}
	p, err = hex.DecodeString(xet.PrefixWrite)
	if err == nil {
		et.PrefixWrite = p
	}

	i, err = strconv.ParseUint(xet.BlockLength, 0, 8)
	if err == nil {
		et.BlockLength = uint8(i)
	}
	i, err = strconv.ParseUint(xet.BlockFactor, 0, 8)
	if err == nil {
		et.BlockFactor = uint8(i)
	}
	i, err = strconv.ParseUint(xet.MappingType, 0, 8)
	if err == nil {
		et.MappingType = uint8(i)
	}
	i, err = strconv.ParseUint(xet.BytePosition, 0, 8)
	if err == nil {
		et.BytePosition = uint8(i)
	}
	i, err = strconv.ParseUint(xet.ByteLength, 0, 8)
	if err == nil {
		et.ByteLength = uint8(i)
	}
	i, err = strconv.ParseUint(xet.BitPosition, 0, 8)
	if err == nil {
		et.BitPosition = uint8(i)
	}
	i, err = strconv.ParseUint(xet.BitLength, 0, 8)
	if err == nil {
		et.BitLength = uint8(i)
	}

	et.ALZ = xet.ALZ

	et.Conversion = xet.Conversion

	f, err := strconv.ParseFloat(xet.ConversionFactor, 32)
	if err == nil {
		et.ConversionFactor = float32(f)
	}
	f, err = strconv.ParseFloat(xet.ConversionOffset, 32)
	if err == nil {
		et.ConversionOffset = float32(f)
	}
	f, err = strconv.ParseFloat(xet.LowerBorder, 32)
	if err == nil {
		et.LowerBorder = float32(f)
	}
	f, err = strconv.ParseFloat(xet.UpperBorder, 32)
	if err == nil {
		et.UpperBorder = float32(f)
	}
	f, err = strconv.ParseFloat(xet.Stepping, 32)
	if err == nil {
		et.Stepping = float32(f)
	}

	et.ValueList = xet.ValueList
	et.Unit = xet.Unit

	return et, err
}

// Conversion is the converter function to use
/*
   "DateBCD"
   "DateMBus"
   "DateTimeBCD"
   "DateTimeMBus"
   "DatenpunktADDR"
   "Div10"
   "Div100"
   "Div1000"
   "Div2"
   "Estrich"
   "HexByte2AsciiByte"
   "HexByte2DecimalByte"
   "HexToFloat"
   "HourDiffSec2Hour"
   "IPAddress"
   "Kesselfolge"
   "Mult10"
   "Mult100"
   "Mult2"
   "Mult5"
   "MultOffset"
   "MultOffsetBCD"
   "MultOffsetFloat"
   "NoConversion"
   "Phone2BCD"
   "RotateBytes"
   "Sec2Hour"
   "Sec2Minute"
   "Time53"
   "UTCDiff2Month"
*/

func str2CmdType(s string) CommandType {
	var c CommandType
	var readWrite byte // 0 == undefined, 1 == read, 2 == write, 3==bidirectional/rpc

	// TODO: find out more CommandType mappings
	switch s {
	case "BE_READ":
		c = nop
		readWrite = 0x01
	case "BE_WRITE":
		c = nop
		readWrite = 0x02
	case "EEPROM_READ":
		c = nop
		readWrite = 0x01
	case "EEPROM_WRITE":
		c = nop
		readWrite = 0x02
	case "KBUS_DATAELEMENT_READ":
		c = nop
		readWrite = 0x01
	case "KBUS_DIRECT_READ":
		c = nop
		readWrite = 0x01
	case "KBUS_DIRECT_WRITE":
		c = nop
		readWrite = 0x02
	case "KBUS_EEPROM_LT_READ":
		c = nop
		readWrite = 0x01
	case "KBUS_EEPROM_LT_WRITE":
		c = nop
		readWrite = 0x02
	case "KBUS_GATEWAY_WRITE":
		c = nop
		readWrite = 0x02
	case "KBUS_INDIRECT_READ":
		c = nop
		readWrite = 0x01
	case "KBUS_INDIRECT_WRITE":
		c = nop
		readWrite = 0x02
	case "KBUS_MEMBERLIST_READ":
		c = nop
		readWrite = 0x01
	case "KBUS_MEMBERLIST_WRITE":
		c = nop
		readWrite = 0x02
	case "KBUS_TRANSPARENT_READ":
		c = nop
		readWrite = 0x01
	case "KBUS_TRANSPARENT_WRITE":
		c = nop
		readWrite = 0x02
	case "KBUS_VIRTUAL_READ":
		c = nop
		readWrite = 0x01
	case "KBUS_VIRTUAL_WRITE":
		c = nop
		readWrite = 0x02
	case "KMBUS_EEPROM_READ":
		c = nop
		readWrite = 0x01
	case "Physical_READ":
		c = nop
		readWrite = 0x01
	case "Port_READ":
		c = nop
		readWrite = 0x01
	case "Remote_Procedure_Call":
		// TODO: Is this p300FunctionCall (0x07)?
		c = nop
		readWrite = 0x00 // TODO: uses 0x07?
	case "Virtual_MBUS":
		c = nop
		readWrite = 0x03
	case "Virtual_MarktManager_READ":
		c = nop
		readWrite = 0x01
	case "Virtual_MarktManager_WRITE":
		c = nop
		readWrite = 0x01
	case "Virtual_READ":
		c = p300ReadData
		readWrite = 0x01
	case "Virtual_WRITE":
		c = p300WriteData
		readWrite = 0x02
	case "Virtual_WILO_READ":
		c = nop
		readWrite = 0x01
	case "Virtual_WILO_WRITE":
		c = nop
		readWrite = 0x02
	case "undefined":
		c = nop
		readWrite = 0x00
	default:
		c = nop
		readWrite = 0x00
	}

	_ = readWrite // TODO: Remove or make use of readWrite
	return c
}

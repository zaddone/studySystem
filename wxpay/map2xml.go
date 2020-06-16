package main
import(
	"encoding/xml"
	"fmt"
	"io"
)
type Map map[string]interface{}

type xmlMapEntry struct {
    XMLName xml.Name
    Value   string `xml:",innerxml"`
}

type xmlMapEntryString struct {
    XMLName xml.Name
    Value   string `xml:",chardata"`
}
type xmlMapEntryInt struct {
    XMLName xml.Name
    Value   int `xml:",chardata"`
}
func (m Map) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
    if len(m) == 0 {
        return nil
    }
    start.Name = xml.Name{Local: "xml"}

    err := e.EncodeToken(start)
    if err != nil {
        return err
    }

    for k, v := range m {
	switch _v := v.(type){
	case string:
	    e.Encode(xmlMapEntryString{XMLName: xml.Name{Local: k}, Value: _v})
	case int:
	    e.Encode(xmlMapEntryInt{XMLName: xml.Name{Local: k}, Value: _v})
	}
    }

    return e.EncodeToken(start.End())
}

func (m *Map) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
    *m = Map{}
    for {
        var e xmlMapEntry
        err := d.Decode(&e)
        if err == io.EOF {
		break
        } else if err != nil {
		fmt.Println(err)
		return err
        }

        (*m)[e.XMLName.Local] = e.Value
    }
    return nil
}
func _main() {
    // The Map
    m := map[string]interface{}{
        "key_1": "Value One",
        "key_2": 1231231,
    }
    fmt.Println(m)
    // Encode to XML
    x, _ := xml.MarshalIndent(Map(m), "", "  ")
    fmt.Println(string(x))

    // Decode back from XML
    var rm map[string]interface{}
    xml.Unmarshal(x, (*Map)(&rm))
    fmt.Println(rm)
}

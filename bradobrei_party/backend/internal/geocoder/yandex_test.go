package geocoder

import (
	"strings"
	"testing"
)

// Фрагмент ответа v1 из документации Яндекса (один объект, found как строка).
const sampleYandexV1 = `{
  "response": {
    "GeoObjectCollection": {
      "metaDataProperty": {
        "GeocoderResponseMetaData": {
          "request": "тест",
          "found": "1",
          "results": "10"
        }
      },
      "featureMember": [
        {
          "GeoObject": {
            "metaDataProperty": {
              "GeocoderMetaData": {
                "text": "Россия, Москва"
              }
            },
            "Point": { "pos": "37.62 55.75" }
          }
        }
      ]
    }
  }
}`

func TestYandexParseGeocodeJSON(t *testing.T) {
	r, err := yandexParseGeocodeJSON([]byte(sampleYandexV1))
	if err != nil {
		t.Fatal(err)
	}
	if r.Lon != 37.62 || r.Lat != 55.75 {
		t.Fatalf("coords: got lat=%v lon=%v", r.Lat, r.Lon)
	}
	if !strings.Contains(r.FormattedAddress, "Москва") {
		t.Fatalf("text: %q", r.FormattedAddress)
	}
}

func TestYandexParseGeocodeJSON_foundAsNumber(t *testing.T) {
	// Некоторые ответы отдают found числом — раньше struct-парсинг мог ломаться на соседних полях.
	const j = `{
  "response": {
    "GeoObjectCollection": {
      "metaDataProperty": {
        "GeocoderResponseMetaData": { "found": 1, "results": 10 }
      },
      "featureMember": [
        {
          "GeoObject": {
            "metaDataProperty": { "GeocoderMetaData": { "text": "X" } },
            "Point": { "pos": "30.3 59.9" }
          }
        }
      ]
    }
  }
}`
	r, err := yandexParseGeocodeJSON([]byte(j))
	if err != nil {
		t.Fatal(err)
	}
	if r.Lon != 30.3 || r.Lat != 59.9 {
		t.Fatalf("coords: got lat=%v lon=%v", r.Lat, r.Lon)
	}
}

func TestYandexParseGeocodeJSON_featureMemberSingleObject(t *testing.T) {
	const j = `{
  "response": {
    "GeoObjectCollection": {
      "metaDataProperty": { "GeocoderResponseMetaData": { "found": "1" } },
      "featureMember": {
        "GeoObject": {
          "metaDataProperty": { "GeocoderMetaData": { "text": "Y" } },
          "Point": { "pos": "60.5 56.8" }
        }
      }
    }
  }
}`
	r, err := yandexParseGeocodeJSON([]byte(j))
	if err != nil {
		t.Fatal(err)
	}
	if r.Lon != 60.5 || r.Lat != 56.8 {
		t.Fatalf("coords: lat=%v lon=%v", r.Lat, r.Lon)
	}
}

func TestYandexParseGeocodeJSON_rootStatusError(t *testing.T) {
	const j = `{"statusCode":403,"message":"Invalid apikey"}`
	_, err := yandexParseGeocodeJSON([]byte(j))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestYandexParseGeocodeJSON_emptyFound(t *testing.T) {
	const j = `{
  "response": {
    "GeoObjectCollection": {
      "metaDataProperty": {
        "GeocoderResponseMetaData": { "found": "0", "results": "10" }
      },
      "featureMember": []
    }
  }
}`
	_, err := yandexParseGeocodeJSON([]byte(j))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "не нашёл") {
		t.Fatalf("unexpected: %v", err)
	}
}

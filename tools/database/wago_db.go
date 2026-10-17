package database

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
)

const (
	itemIDHeader  = "ID"
	flags1Header  = "Flags_1"
	itemSetHeader = "ItemSet"

	flag1AllianceOnly = 0x6002
	flag1HordeOnly    = 0x6001
)

type WagoDbItem struct {
	FactionRestriction proto.UIItem_FactionRestriction
	ItemSetID          int32
}

func flags1ToFactionRestriction(flags1 int) proto.UIItem_FactionRestriction {
	switch flags1 {
	case flag1AllianceOnly:
		return proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY
	case flag1HordeOnly:
		return proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY
	default:
		return proto.UIItem_FACTION_RESTRICTION_UNSPECIFIED
	}
}

func ParseWagoDB(dbContents string) map[int32]WagoDbItem {
	r := csv.NewReader(strings.NewReader(dbContents))
	rawHeaders, err := r.Read()
	if err != nil {
		log.Fatalf("Cannot read wago csv header row: %v", err)
	}

	headerMap := map[string]int{}
	for i, name := range rawHeaders {
		headerMap[name] = i
	}

	if _, ok := headerMap[itemIDHeader]; !ok {
		log.Fatalf("The wago csv does not have an ID header column. All columns: %#v", headerMap)
	}
	if _, ok := headerMap[flags1Header]; !ok {
		log.Fatalf("The wago csv does not have a Flags_1 header column. All columns: %#v", headerMap)
	}
	if _, ok := headerMap[itemSetHeader]; !ok {
		log.Fatalf("The wago csv does not have a ItemSet header column. All columns: %#v", headerMap)
	}

	result := map[int32]WagoDbItem{}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Cannot read wago csv row: %v", err)
		}

		itemID, err := strconv.Atoi(row[headerMap[itemIDHeader]])
		if err != nil {
			log.Fatalf("Cannot parse ItemID from row %v: %v", row, err)
		}

		flags1, err := strconv.Atoi(row[headerMap[flags1Header]])
		if err != nil {
			log.Fatalf("Cannot parse Flags_1 from row %v: %v", row, err)
		}

		itemSet, err := strconv.Atoi(row[headerMap[itemSetHeader]])
		if err != nil {
			log.Fatalf("Cannot parse ItemSet from row %v: %v", row, err)
		}

		result[int32(itemID)] = WagoDbItem{
			FactionRestriction: flags1ToFactionRestriction(flags1),
			ItemSetID:          int32(itemSet),
		}
	}

	return result
}

// OverlayForeverClientItems lays the Forever client's item facts over Classic's.
//
// wagoItems is Classic 1.15.3's ItemSparse (wago_db2_items.csv), and contents is
// forever_client_items.csv from tools/sources/itemsparse.py. An item in the Forever file
// takes its set from there, and its faction too when the file has its ItemSparse row
// (Flags_1 is blank for an item only the ItemSet table lists). Every other item keeps
// Classic's, because the Forever ItemSparse on wago.tools lacks most Classic items.
func OverlayForeverClientItems(wagoItems map[int32]WagoDbItem, contents string) {
	r := csv.NewReader(strings.NewReader(contents))
	header, err := r.Read()
	if err != nil {
		log.Fatalf("Cannot read forever_client_items.csv header row: %v", err)
	}
	col := map[string]int{}
	for i, name := range header {
		col[name] = i
	}
	for _, name := range []string{itemIDHeader, flags1Header, itemSetHeader} {
		if _, ok := col[name]; !ok {
			log.Fatalf("forever_client_items.csv has no %s column. All columns: %#v", name, header)
		}
	}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Cannot read forever_client_items.csv row: %v", err)
		}
		itemID, err := strconv.Atoi(row[col[itemIDHeader]])
		if err != nil {
			log.Fatalf("Cannot parse ID from row %v: %v", row, err)
		}
		item := wagoItems[int32(itemID)]
		if flags := row[col[flags1Header]]; flags != "" {
			flags1, err := strconv.Atoi(flags)
			if err != nil {
				log.Fatalf("Cannot parse Flags_1 from row %v: %v", row, err)
			}
			item.FactionRestriction = flags1ToFactionRestriction(flags1)
		}
		itemSet, err := strconv.Atoi(row[col[itemSetHeader]])
		if err != nil {
			log.Fatalf("Cannot parse ItemSet from row %v: %v", row, err)
		}
		item.ItemSetID = int32(itemSet)
		wagoItems[int32(itemID)] = item
	}
}

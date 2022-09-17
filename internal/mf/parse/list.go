package parse

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tubo28/moneyforward-scraper/internal/mf"
)

func List(html []byte, year, month int) ([]*mf.MFTransaction, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("cannot build goquery Document: %w", err)
	}

	var ret []*mf.MFTransaction

	doc.Find(".list_body .transaction_list").Each(func(i int, s *goquery.Selection) {
		var tds []*goquery.Selection
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			tds = append(tds, s)
		})

		if len(tds) < 8 {
			return
		}

		transactionID, _ := tds[0].Find("input#user_asset_act_id").First().Attr("value")
		date, err := strconv.Atoi(strings.TrimSpace(tds[1].Text())[3:5])
		if err != nil {
			log.Print("failed to parse transactionID field: %w", err)
			return
		}
		amount, err := strconv.Atoi(strings.ReplaceAll(strings.TrimSpace(tds[3].Find("span").Text()), ",", ""))
		if err != nil {
			log.Print("failed to parse amount field: %w", err)
			return
		}
		institution, _ := tds[4].Attr("title")

		mft := &mf.MFTransaction{
			TransactionID:     transactionID,
			IsCalculateTarget: tds[0].Find("i.icon-check").Length() != 0,
			Date:              time.Date(year, time.Month(month), date, 0, 0, 0, 0, time.Local),
			Content:           strings.TrimSpace(tds[2].Text()),
			Amount:            amount,
			Institution:       strings.TrimSpace(institution),
			LargeCategory:     strings.TrimSpace(tds[5].Text()),
			MiddleCategory:    strings.TrimSpace(tds[6].Text()),
			Memo:              strings.TrimSpace(tds[7].Text()),
		}
		ret = append(ret, mft)

		log.Printf("parsed id:%s date:%s", mft.TransactionID, mft.Date)
	})

	return ret, nil
}

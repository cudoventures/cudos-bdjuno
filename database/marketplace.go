package database

import (
	"fmt"
	"time"

	"github.com/forbole/bdjuno/v2/database/utils"
)

func (db *Db) CheckIfNftExists(tokenID uint64, denomID string) error {
	var rows []string

	err := db.Sqlx.Select(&rows, `SELECT denom_id FROM marketplace_nft WHERE token_id=$1 AND denom_id=$2`, tokenID, denomID)
	if err != nil {
		return err
	}

	if len(rows) != 1 {
		return fmt.Errorf("not found")
	}

	return nil
}

func (db *Db) SaveMarketplaceCollection(txHash string, id uint64, denomID, mintRoyalties, resaleRoyalties, creator string, verified bool) error {
	_, err := db.Sql.Exec(`INSERT INTO marketplace_collection (transaction_hash, id, denom_id, mint_royalties, resale_royalties, verified, creator) 
		VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`, txHash, id, denomID, mintRoyalties, resaleRoyalties, verified, creator)
	return err
}

func (tx *DbTx) ListNft(txHash string, id, tokenID uint64, denomID, price string) error {
	_, err := tx.Exec(`UPDATE marketplace_nft SET transaction_hash=$1, id=$2, price=$3 WHERE token_id=$4 AND denom_id=$5`,
		txHash, id, price, tokenID, denomID)
	fmt.Println(err)
	return err
}

func (tx *DbTx) SaveMarketplaceNft(txHash string, tokenID uint64, denomID, uid, price, creator string) error {
	_, err := tx.Exec(`INSERT INTO marketplace_nft (transaction_hash, uid, token_id, denom_id, price, creator, uniq_id) 
		VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (token_id, denom_id) DO UPDATE SET price = EXCLUDED.price, id = EXCLUDED.id`,
		txHash, uid, tokenID, denomID, price, creator, utils.FormatUniqID(tokenID, denomID))
	return err
}

func (tx *DbTx) SaveMarketplaceNftBuy(txHash string, id uint64, buyer string, timestamp uint64, usdPrice, btcPrice string) error {
	var tokenID uint64
	var denomID, price, seller string

	if err := tx.QueryRow(`SELECT token_id, denom_id, price, creator FROM marketplace_nft WHERE id = $1`, id).Scan(&tokenID, &denomID, &price, &seller); err != nil {
		return err
	}

	if seller == "" {
		return fmt.Errorf("nft (%d) not found for sale", id)
	}

	return tx.saveMarketplaceNftBuy(txHash, buyer, timestamp, tokenID, denomID, price, seller, usdPrice, btcPrice)
}

func (tx *DbTx) saveMarketplaceNftBuy(txHash string, buyer string, timestamp, tokenID uint64, denomID, price, seller, usdPrice, btcPrice string) error {
	_, err := tx.Exec(`INSERT INTO marketplace_nft_buy_history (transaction_hash, token_id, denom_id, price, seller, buyer, usd_price, btc_price, timestamp, uniq_id) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, txHash, tokenID, denomID, price, seller, buyer, usdPrice, btcPrice, timestamp, utils.FormatUniqID(tokenID, denomID))
	return err
}

func (tx *DbTx) SaveMarketplaceNftMint(txHash string, tokenID uint64, buyer, denomID, price string, timestamp uint64, usdPrice, btcPrice string) error {
	return tx.saveMarketplaceNftBuy(txHash, buyer, timestamp, tokenID, denomID, price, "0x0", usdPrice, btcPrice)
}

func (tx *DbTx) SetMarketplaceNFTPrice(id uint64, price string) error {
	_, err := tx.Exec(`UPDATE marketplace_nft SET price = $1 WHERE id = $2`, price, id)
	return err
}

func (tx *DbTx) UnlistNft(id uint64) error {
	_, err := tx.Exec(`UPDATE marketplace_nft SET price = '0', id = null WHERE id = $1`, id)
	return err
}

func (tx *DbTx) SaveMarketplaceAuction(auctionID uint64, tokenID uint64, denomID string, creator string, startTime time.Time, endTime time.Time, auctionInfo string) error {
	_, err := tx.Exec(`INSERT INTO marketplace_auction (id, token_id, denom_id, creator, start_time, end_time, auction)
	VALUES($1, $2, $3, $4, $5, $6, $7)`, auctionID, tokenID, denomID, creator, startTime, endTime, auctionInfo)
	return err
}

func (tx *DbTx) SaveMarketplaceBid(auctionID uint64, bidder string, price string, timestamp time.Time, txHash string) error {
	_, err := tx.Exec(`INSERT INTO marketplace_bid (auction_id, bidder, price, timestamp, transaction_hash)
	VALUES($1, $2, $3, $4, $5)`, auctionID, bidder, price, timestamp, txHash)
	return err
}

func (tx *DbTx) SaveMarketplaceAuctionSold(auctionID uint64, timestamp uint64, usdPrice string, btcPrice string, txHashAcceptBid string) error {
	var tokenID uint64
	var denomID, seller, buyer, price, txHashPlaceBid string

	if err := tx.QueryRow(`SELECT token_id, denom_id, creator FROM marketplace_auction WHERE id = $1`, auctionID).Scan(&tokenID, &denomID, &seller); err != nil {
		return err
	}

	if err := tx.QueryRow(`SELECT bidder, transaction_hash, price FROM marketplace_bid WHERE auction_id = $1 ORDER BY timestamp DESC`, auctionID).Scan(&buyer, &txHashPlaceBid, &price); err != nil {
		return err
	}

	_, err := tx.Exec(`UPDATE marketplace_auction SET sold = true WHERE id = $1`, auctionID)
	if err != nil {
		return err
	}

	txHashBuyNft := txHashAcceptBid
	if txHashBuyNft == "" {
		txHashBuyNft = txHashPlaceBid
	}

	if err := tx.saveMarketplaceNftBuy(txHashBuyNft, buyer, timestamp, tokenID, denomID, price, seller, usdPrice, btcPrice); err != nil {
		return err
	}

	return tx.UpdateNFTHistory(txHashBuyNft, tokenID, denomID, seller, buyer, timestamp)
}

func (tx *DbTx) UpdateMarketplaceAuctionInfo(auctionID uint64, auctionInfo string) error {
	_, err := tx.Exec(`UPDATE marketplace_auction SET auction = $2 WHERE id = $1`, auctionID, auctionInfo)
	return err
}

func (db *Db) UnlistNft(id uint64) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_nft SET price = '0', id = null WHERE id = $1`, id)
	return err
}

func (db *Db) SetMarketplaceCollectionVerificationStatus(id uint64, verified bool) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_collection SET verified = $1 WHERE id = $2`, verified, id)
	return err
}

func (db *Db) SetMarketplaceNFTPrice(id uint64, price string) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_nft SET price = $1 WHERE id = $2`, price, id)
	return err
}

func (db *Db) SetMarketplaceCollectionRoyalties(id uint64, mintRoyalties, resaleRoyalties string) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_collection SET mint_royalties = $1, resale_royalties = $2 WHERE id = $3`, mintRoyalties, resaleRoyalties, id)
	return err
}

func (db *Db) UpdateMarketplaceAuctionInfo(auctionID uint64, auctionInfo string) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_auction SET auction = $2 WHERE id = $1`, auctionID, auctionInfo)
	return err
}

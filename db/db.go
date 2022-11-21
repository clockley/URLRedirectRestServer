package db

import (
	"context"
	"database/sql"
	"dwarfRestServer/hash"
	"encoding/json"
	"fmt"
	"github.com/creasty/defaults"
	_ "github.com/microsoft/go-mssqldb/azuread"
	"net/url"
)

const retryCount = 5

type MyNullString struct {
	sql.NullString
}

func (s MyNullString) MarshalJSON() ([]byte, error) {
	if s.Valid {
		return json.Marshal(s.String)
	}
	return []byte(`null`), nil
}

type Payload struct {
	Url          string `json:"originalURL"`
	ShortUrl     string `json:"shortURL"`
	Title        string `json:"title"`
	Domain       string `default:"dwarf.me"`
	RedirectType int    `json:"redirectType" default:"302"`
}

type HashInfo struct {
	Id           int64 `default:"-1"`
	UserId       int64 `default:"-1"`
	HashMethod   string
	HashId       string
	Domain       string
	Salt         string `json:"-"`
	Title        MyNullString
	TargetUrl    string
	ExpiredUrl   MyNullString
	DateCreated  string
	ExpireTime   MyNullString
	RedirectType int `default:"302"`
}

type DatabaseConnection struct {
	Con             *sql.DB
	LookupUrlStmt   *sql.Stmt
	GetHashInfoStmt *sql.Stmt
}

func (dbinst DatabaseConnection) Close() {
	dbinst.LookupUrlStmt.Close()
	dbinst.GetHashInfoStmt.Close()
	println("database closed")
}

func (dbinst DatabaseConnection) ConnectToDb() *DatabaseConnection {
	query := url.Values{}
	query.Add("app name", "DwarfServer")

	server := ""
	user := ""
	password := ""
	port := 1433
	database := ""
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;encrypt=True;tlsmin=1.2",
		server, user, password, port, database)
	var err error
	for j := 0; j < retryCount; j++ {
		dbinst.Con, err = sql.Open("sqlserver", connString)
		if err == nil {
			break
		}
	}

	if err != nil {
		panic(err)
	}

	HashInfoQueryString := "SELECT Id, UserId, HashMethod, HashId, DomainName, Salt, Title, TargetUrl, ExpiredUrl, DateCreated FROM UrlTable WHERE HashId = @p1;"
	dbinst.GetHashInfoStmt, err = dbinst.Con.Prepare(HashInfoQueryString)

	if err != nil {
		fmt.Println(err.Error())
	}

	lookupUrl := "SELECT Id, UserId, HashMethod, HashId, DomainName, Salt, Title, TargetUrl, ExpiredUrl, DateCreated, ExpireTime, RedirectType from UrlTable where TargetUrl = @p1"
	dbinst.LookupUrlStmt, err = dbinst.Con.Prepare(lookupUrl)
	if err != nil {
		println(err.Error())
	}

	return &dbinst
}

func (dbinst DatabaseConnection) LookupURL(url string) *HashInfo {
	r := HashInfo{}
	if err := defaults.Set(&r); err != nil {
		panic(err)
	}

	var err error

	err = dbinst.Con.PingContext(context.Background())
	if err != nil {
		fmt.Println(err.Error())
	}

	err = dbinst.LookupUrlStmt.QueryRow(url).Scan(&r.Id, &r.UserId, &r.HashMethod, &r.HashId, &r.Domain, &r.Salt,
		&r.Title, &r.TargetUrl, &r.ExpiredUrl, &r.DateCreated, &r.ExpireTime, &r.RedirectType)
	if err != nil {
		println(err.Error())
		if err := defaults.Set(&r); err != nil {
			panic(err)
		}
		return &r
	}
	return &r
}

func (dbinst DatabaseConnection) GetHashInfo(hash string) *HashInfo {
	r := HashInfo{}
	if err := defaults.Set(&r); err != nil {
		panic(err)
	}

	ctx := context.Background()
	err := dbinst.Con.PingContext(ctx)

	err = dbinst.GetHashInfoStmt.QueryRow(hash).Scan(&r.Id, &r.UserId, &r.HashMethod, &r.HashId, &r.Domain, &r.Salt, &r.Title, &r.TargetUrl, &r.ExpiredUrl, &r.DateCreated)
	if err != nil {
		println(err.Error())
		return nil
	}
	return &r
}

func (dbinst DatabaseConnection) CreateShortUrl(i *Payload) {
	ctx := context.Background()
	err := dbinst.Con.PingContext(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
	var rowId int64 = -1

	if err != nil {
		panic(err.Error())
	}

	salt := hash.RandStringBytesMaskImprSrcSB(16)

	for j := 0; j < retryCount; j++ {
		row := dbinst.Con.QueryRowContext(ctx,
			"INSERT INTO dbo.UrlTable (HashMethod, TargetUrl, Salt) VALUES (@p1, @p2, @p3);"+
				"select ID = convert(bigint, SCOPE_IDENTITY());",
			"hashid", i.Url, salt)

		if row.Scan(&rowId) == nil {
			break
		}
		fmt.Println("retrying insert")
	}

	hashId := hash.CreateHash("", salt, rowId)

	dbinst.Con.Exec("UPDATE UrlTable SET HashId = @p1 WHERE Id = @p2", hashId, rowId)
	i.ShortUrl = fmt.Sprintf("https://www.%s/%s", i.Domain, hashId)

	if err != nil {
		panic(err.Error())
	}
}

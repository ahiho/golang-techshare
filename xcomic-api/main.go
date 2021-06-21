package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB = nil
)

type BaseModel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
}

type Comic struct {
	BaseModel
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Chapters    []Chapter `json:"chapters"`
}

type Chapter struct {
	BaseModel
	Number  uint          `json:"number"`
	Title   string        `json:"title"`
	ComicID uint          `json:"-"`
	Pages   []ChapterPage `json:"pages"`
}

type ChapterPage struct {
	BaseModel
	Number    uint   `json:"number"`
	Url       string `json:"url"`
	ChapterID uint   `json:"-"`
}

func main() {
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)
	_db, err := gorm.Open(sqlite.Open("xcomic.db"), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Printf("Cannot connect to database %v", err)
		return
	}
	_db.AutoMigrate(&Comic{}, &Chapter{}, &ChapterPage{})

	db = _db
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/heathz", heathz).Methods("GET")

	router.HandleFunc("/comics", searchComic).Methods("GET")
	router.HandleFunc("/comics", createNewComic).Methods("POST")

	router.HandleFunc("/comics/{id}", getComic).Methods("GET")
	router.HandleFunc("/comics/{id}", editComic).Methods("PUT")
	// router.HandleFunc("/comics/{id}", deleteComic).Methods("DELETE")

	router.HandleFunc("/comics/{id}/chapters", listChapers).Methods("GET")
	router.HandleFunc("/comics/{id}/chapters", addChapter).Methods("POST")

	router.HandleFunc("/chapter/{id}", getChapter).Methods("GET")
	router.HandleFunc("/chapter/{id}", editChapter).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8080", router))

}

func heathz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func responseError(w http.ResponseWriter, e error) {
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"error": e.Error(),
	})
}

func readRequestBody(r *http.Request) {

}

func responseSuccess(w http.ResponseWriter, v interface{}) {
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func searchComic(w http.ResponseWriter, r *http.Request) {
	comics := []Comic{}
	urlParams := r.URL.Query()
	keyword := urlParams.Get("q")
	_db := db
	if keyword != "" {
		_db = _db.Where("title like ?", "%"+keyword+"%")
	}
	_db.Limit(10).Find(&comics)
	responseSuccess(w, comics)
}

func getComic(w http.ResponseWriter, r *http.Request) {
	comic := Comic{}
	vars := mux.Vars(r)
	id := vars["id"]
	idNum, err := strconv.Atoi(id)
	if err != nil {
		responseError(w, err)
		return
	}
	err = db.Where("id = ?", idNum).First(&comic).Error
	if err != nil {
		responseError(w, err)
		return
	}
	responseSuccess(w, comic)
}

func getChapter(w http.ResponseWriter, r *http.Request) {
	chapter := Chapter{}
	vars := mux.Vars(r)
	id := vars["id"]
	idNum, err := strconv.Atoi(id)
	if err != nil {
		responseError(w, err)
		return
	}
	err = db.Preload("chapter_pages").Where("id = ?", idNum).First(&chapter).Error
	if err != nil {
		responseError(w, err)
		return
	}
	responseSuccess(w, chapter)
}

func createNewComic(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(w, err)
		return
	}
	var comic Comic
	err = json.Unmarshal(body, &comic)
	if err != nil {
		responseError(w, err)
		return
	}
	err = db.Create(&comic).Error
	if err != nil {
		responseError(w, err)
		return
	}
	responseSuccess(w, comic)
}

func editComic(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(w, err)
		return
	}
	var updateComic Comic
	err = json.Unmarshal(body, &updateComic)
	if err != nil {
		responseError(w, err)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	idNum, err := strconv.Atoi(id)
	if err != nil {
		responseError(w, err)
		return
	}
	var comic Comic
	err = db.Where("id = ?", idNum).First(&comic).Error
	if err != nil {
		responseError(w, err)
		return
	}
	responseSuccess(w, comic)
}

func deleteComic(w http.ResponseWriter, r *http.Request) {
	responseSuccess(w, map[string]string{
		"error": "Not impelemeted",
	})
}

func listChapers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idNum, err := strconv.Atoi(id)
	if err != nil {
		responseError(w, err)
		return
	}

	var chapters []Chapter
	err = db.Where("comic_id = ?", idNum).Order("number asc").Find(&chapters).Error
	if err != nil {
		responseError(w, err)
		return
	}
	responseSuccess(w, chapters)
}

func addChapter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idNum, err := strconv.Atoi(id)
	if err != nil {
		responseError(w, err)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(w, err)
		return
	}
	var newChapter Chapter
	err = json.Unmarshal(body, &newChapter)
	if err != nil {
		responseError(w, err)
		return
	}
	var chapter Chapter
	chapter.ComicID = uint(idNum)
	err = db.Create(&chapter).Error
	if err != nil {
		responseError(w, err)
		return
	}
	pages := newChapter.Pages
	if len(pages) > 0 {
		for _, page := range pages {
			page.ChapterID = chapter.ID
		}
		db.Create(&pages)
	}
	chapter.Pages = pages
	responseSuccess(w, chapter)
}

func editChapter(w http.ResponseWriter, r *http.Request) {
	responseSuccess(w, map[string]string{
		"error": "Not impelemeted",
	})
}

func deleteChapter(w http.ResponseWriter, r *http.Request) {
	responseSuccess(w, map[string]string{
		"error": "Not impelemeted",
	})
}

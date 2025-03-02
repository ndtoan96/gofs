package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"

	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"

	"strings"

	"github.com/ndtoan96/gofs/model"
	cp "github.com/otiai10/copy"
	"github.com/spf13/pflag"

	_ "embed"
)

// static files

//go:embed templates/layout.html
var htmlLayout string

//go:embed templates/delete.html
var htmlDelete string

//go:embed templates/new-folder.html
var htmlNewFolder string

//go:embed templates/archive.html
var htmlArchive string

//go:embed templates/files.html
var htmlFiles string

//go:embed templates/upload.html
var htmlUpload string

//go:embed templates/rename.html
var htmlRename string

//go:embed templates/edit.html
var htmlEdit string

//go:embed templates/search.html
var htmlSearch string

//go:embed static/style.css
var cssStyle string

//go:embed static/favicon.svg
var faviconImg []byte

//go:embed static/script.js
var scriptSource string

func servePath(w http.ResponseWriter, r *http.Request) {
	p := r.PathValue("path")
	selectState := r.URL.Query().Get("select")
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") || path.IsAbs(p) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if p == "" {
		p = "."
	}
	sortField := r.URL.Query().Get("sort")
	if sortField == "" || (sortField != "mtime" && sortField != "size") {
		sortField = "name"
	}
	desc := r.URL.Query().Get("desc")
	var descb bool
	if desc != "true" {
		descb = false
	} else {
		descb = true
	}
	preview := r.URL.Query().Get("preview")
	htmlPreview := getHtmlPreview(p, preview)

	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if info.IsDir() {
		dir, err := os.ReadDir(p)
		if err != nil {
			log.Println("[ERROR]", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Serve directory
		items := make([]model.Item, 0)
		for _, e := range dir {
			info, _ := e.Info()
			if e.IsDir() {
				d, _ := os.ReadDir(e.Name())
				items = append(items, model.Item{IsDir: true, Name: e.Name(), LastModified: info.ModTime(), Size: model.DirSize(len(d))})
			} else {
				if ext := path.Ext(e.Name()); ext == ".zip" || ext == ".7z" || ext == ".gz" || ext == ".rar" || ext == ".tar" || ext == ".jar" || ext == ".bz" || ext == ".epub" {
					items = append(items, model.Item{IsDir: false, Name: e.Name(), LastModified: info.ModTime(), Size: model.FileSize(info.Size()), FileType: "archive"})
				} else if ext := path.Ext(e.Name()); ext == ".jpg" || ext == ".png" || ext == ".gif" || ext == ".webp" || ext == ".jpeg" || ext == ".svg" {
					items = append(items, model.Item{IsDir: false, Name: e.Name(), LastModified: info.ModTime(), Size: model.FileSize(info.Size()), FileType: "image"})
				} else if yes, _ := isTextFile(path.Join(p, e.Name())); yes {
					items = append(items, model.Item{IsDir: false, Name: e.Name(), LastModified: info.ModTime(), Size: model.FileSize(info.Size()), FileType: "text"})
				} else {
					items = append(items, model.Item{IsDir: false, Name: e.Name(), LastModified: info.ModTime(), Size: model.FileSize(info.Size()), FileType: "binary"})
				}
			}
		}
		sort.Slice(items, func(i, j int) bool {
			var first int
			var second int
			if !descb {
				first = i
				second = j
			} else {
				first = j
				second = i
			}
			switch sortField {
			case "mtime":
				return items[first].LastModified.Before(items[second].LastModified)
			case "size":
				if items[i].IsDir && !items[j].IsDir {
					return true
				}
				if !items[i].IsDir && items[j].IsDir {
					return false
				}
				if items[first].IsDir && items[second].IsDir {
					return items[first].Size.(model.DirSize) < items[second].Size.(model.DirSize)
				} else {
					return items[first].Size.(model.FileSize) < items[second].Size.(model.FileSize)
				}
			default:
				return strings.ToLower(items[first].Name) < strings.ToLower(items[second].Name)
			}
		})
		if err := tmpl["files"].Execute(w, model.FilesPageModel{Path: model.Path(p), Items: items, AllowWrite: allowWrite, SelectState: selectState, SortField: sortField, Desc: descb, Preview: htmlPreview}); err != nil {
			log.Fatalln("[ERROR]", err)
		}

	} else {
		http.ServeFile(w, r, p)
	}
}

func action(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	names := r.PostForm["select"]
	p := r.FormValue("path")
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") || path.IsAbs(p) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	var err error = nil
	action := r.FormValue("action")
	switch action {
	case "delete":
		if len(names) > 0 {
			err = tmpl["delete"].Execute(w, model.DeletePageModel{Path: model.Path(p), Names: names})
		} else {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
		}
	case "archive":
		if len(names) > 0 {
			err = tmpl["archive"].Execute(w, model.ArchivePageModel{Path: model.Path(p), Names: names})
		} else {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
		}
	case "new-folder":
		err = tmpl["new-folder"].Execute(w, model.NewFolderPageModel{Path: model.Path(p)})
	case "upload":
		err = tmpl["upload"].Execute(w, model.NewFolderPageModel{Path: model.Path(p)})
	case "rename":
		if len(names) > 0 {
			err = tmpl["rename"].Execute(w, model.RenamePageModel{Path: model.Path(p), OldNames: names})
		} else {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
		}
	case "edit":
		editableFiles := make([]string, 0)
		for _, name := range names {
			yes, _ := isTextFile(path.Join(p, name))
			if yes {
				editableFiles = append(editableFiles, name)
			}
		}
		if len(editableFiles) > 0 {
			contents := make([]string, 0)
			for _, f := range editableFiles {
				content, tmp_err := os.ReadFile(path.Join(p, f))
				if tmp_err != nil {
					log.Println("[ERROR]", tmp_err)
				}
				contents = append(contents, string(content))
			}
			err = tmpl["edit"].Execute(w, model.EditPageModel{Path: model.Path(p), Names: editableFiles, Contents: contents})
		} else {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
		}
	case "download":
		if len(names) == 0 {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
		} else if len(names) == 1 {
			item := names[0]
			itemPath := path.Join(p, item)
			var stat os.FileInfo
			stat, err = os.Stat(itemPath)
			if err != nil {
				break
			}
			if stat.IsDir() {
				var buf bytes.Buffer
				err = zipFilesAndFolders(&buf, p, names)
				if err != nil {
					break
				}
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v.zip", item))
				w.Header().Set("Content-Type", "application/zip")
				io.Copy(w, &buf)
			} else {
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v", item))
				http.ServeFile(w, r, itemPath)
			}
		} else {
			var buf bytes.Buffer
			err = zipFilesAndFolders(&buf, p, names)
			if err != nil {
				break
			}
			zipName := path.Base(p)
			if zipName == "" || zipName == "." {
				zipName = "download"
			}
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%v.zip", zipName))
			w.Header().Set("Content-Type", "application/zip")
			io.Copy(w, &buf)
		}
	case "copy":
		if len(names) > 0 {
			http.SetCookie(w, &http.Cookie{Name: "cpdir", Value: url.QueryEscape(p)})
			encodedNames := strings.Join(names, "!$!")
			http.SetCookie(w, &http.Cookie{Name: "cpitems", Value: url.QueryEscape(encodedNames)})
			http.SetCookie(w, &http.Cookie{Name: "delorigin", Value: "false"})
		}
		http.Redirect(w, r, p, http.StatusMovedPermanently)
	case "cut":
		if len(names) > 0 {
			http.SetCookie(w, &http.Cookie{Name: "cpdir", Value: url.QueryEscape(p)})
			encodedNames := strings.Join(names, "!$!")
			http.SetCookie(w, &http.Cookie{Name: "cpitems", Value: url.QueryEscape(encodedNames)})
			http.SetCookie(w, &http.Cookie{Name: "delorigin", Value: "true"})
		}
		http.Redirect(w, r, p, http.StatusMovedPermanently)
	case "paste":
		var cpdir *http.Cookie
		cpdir, err = r.Cookie("cpdir")
		if err != nil {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
			return
		}
		cpdirValue, err := url.QueryUnescape(cpdir.Value)
		if err != nil {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
		}
		var encodedNames *http.Cookie
		encodedNames, err = r.Cookie("cpitems")
		if err != nil {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
			return
		}
		decodedNamesValue, err := url.QueryUnescape(encodedNames.Value)
		if err != nil {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
			return
		}
		var delOrigin *http.Cookie
		delOrigin, err = r.Cookie("delorigin")
		if err != nil {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
			return
		}
		decodedNames := strings.Split(decodedNamesValue, "!$!")
		for _, name := range decodedNames {
			srcPath := path.Join(cpdirValue, name)
			destPath := path.Join(p, name)
			_, err = os.Stat(destPath)
			for !os.IsNotExist(err) {
				ext := path.Ext(destPath)
				destPath = path.Join(p, strings.TrimSuffix(path.Base(destPath), ext)+"_copy"+ext)
				_, err = os.Stat(destPath)
			}
			err = nil
			// Cannot copy a folder into itself
			if !strings.HasPrefix(destPath, srcPath) {
				cp.Copy(srcPath, destPath)
				if delOrigin.Value == "true" {
					os.RemoveAll(srcPath)
				}
			}
		}
		http.Redirect(w, r, p, http.StatusMovedPermanently)
	case "search":
		text := r.FormValue("search")
		if text == "" {
			http.Redirect(w, r, p, http.StatusMovedPermanently)
			return
		}
		var results []model.SearchResult
		results, err = doSearch(p, text)
		if err == nil {
			err = tmpl["search"].Execute(w, model.SearchPageModel{Search: text, Results: results, Path: model.Path(p)})
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
	if err != nil {
		log.Println("[ERROR]", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	p := r.FormValue("path")
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") || path.IsAbs(p) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.FormValue("submit") == "Yes" {
		items := r.PostForm["items"]
		for _, item := range items {
			deletePath := path.Join(p, item)
			log.Printf("Delete `%v`\n", deletePath)
			err := os.RemoveAll(deletePath)
			if err != nil {
				log.Println("[ERROR]", err)
			}
		}
	}
	http.Redirect(w, r, p, http.StatusMovedPermanently)
}

func newFolder(w http.ResponseWriter, r *http.Request) {
	p := r.FormValue("path")
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") || path.IsAbs(p) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	name := r.FormValue("name")
	if r.FormValue("submit") == "Create" {
		os.Mkdir(path.Join(p, name), 0666)
	}
	http.Redirect(w, r, p, http.StatusMovedPermanently)
}

func archive(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	p := r.FormValue("path")
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") || path.IsAbs(p) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	name := r.FormValue("name")
	name += ".zip"
	if r.FormValue("submit") == "Archive" {
		items := r.PostForm["items"]
		archive, err := os.Create(path.Join(p, name))
		if err != nil {
			log.Println("[ERROR]", err)
		}
		defer archive.Close()
		err = zipFilesAndFolders(archive, p, items)
		if err != nil {
			log.Println("[ERROR]", err)
		}
	}
	http.Redirect(w, r, p, http.StatusMovedPermanently)
}

func zipFilesAndFolders(writer io.Writer, dir string, items []string) error {
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()
	for len(items) > 0 {
		item := items[0]
		items = items[1:]
		item_path := path.Join(dir, item)
		info, err := os.Stat(item_path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirEntry, err := os.ReadDir(item_path)
			if err != nil {
				return err
			}
			if len(dirEntry) == 0 {
				if _, err := zipWriter.Create(item + "/"); err != nil {
					return err
				}
			} else {
				for _, sub := range dirEntry {
					items = append(items, path.Join(item, sub.Name()))
				}
			}

		} else {
			f, err := os.Open(item_path)
			if err != nil {
				return err
			}
			defer f.Close()
			w, err := zipWriter.Create(item)
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, f); err != nil {
				return err
			}
		}
	}
	return nil
}

func rename(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	p := r.FormValue("path")
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") || path.IsAbs(p) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	oldNames := make([]string, 0)
	newNames := make([]string, 0)
	for k, v := range r.PostForm {
		if strings.HasPrefix(k, "oldname-") {
			oldNames = append(oldNames, strings.TrimPrefix(k, "oldname-"))
			newNames = append(newNames, v[0])
		}
	}
	if r.FormValue("submit") == "Rename" {
		for i, oldName := range oldNames {
			newName := newNames[i]
			oldNamePath := path.Join(p, oldName)
			newNamePath := path.Join(p, newName)
			log.Printf("[INFO] Rename `%v` -> `%v`", oldNamePath, newNamePath)
			if err := os.Rename(oldNamePath, newNamePath); err != nil {
				log.Println("[ERROR]", err)
			}
		}
	}
	http.Redirect(w, r, p, http.StatusMovedPermanently)
}

func upload(w http.ResponseWriter, r *http.Request) {
	p := r.FormValue("path")
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") || path.IsAbs(p) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		os.MkdirAll(p, 0666)
	}
	if r.FormValue("submit") == "Upload" {
		for _, f := range r.MultipartForm.File["files"] {
			filePath := path.Join(p, f.Filename)
			_, err := os.Stat(filePath)
			for !os.IsNotExist(err) {
				filePath = path.Join(p, strings.TrimSuffix(path.Base(filePath), path.Ext(filePath))+"_new"+path.Ext(filePath))
				_, err = os.Stat(filePath)
			}
			log.Printf("[INFO] Upload `%v`\n", filePath)
			w, err := os.Create(filePath)
			if err != nil {
				log.Println("[ERROR]", err)
			}
			defer w.Close()
			r, err := f.Open()
			if err != nil {
				log.Println("[ERROR]", err)
			}
			defer r.Close()
			_, err = io.Copy(w, r)
			if err != nil {
				log.Println("[ERROR]", err)
			}
		}
	}
	http.Redirect(w, r, p, http.StatusMovedPermanently)
}

func edit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	p := r.FormValue("path")
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") || path.IsAbs(p) {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.FormValue("submit") == "Save" {
		for k, v := range r.PostForm {
			if strings.HasPrefix(k, "content-") {
				name := strings.TrimPrefix(k, "content-")
				os.WriteFile(path.Join(p, name), []byte(v[0]), 0666)
			}
		}
	}
	http.Redirect(w, r, p, http.StatusMovedPermanently)
}

// global vars
var tmpl map[string]*template.Template
var allowWrite bool

func main() {
	var port int
	var workingDir string
	var host string
	var tlsCert string
	var tlsKey string
	var index bool
	pflag.BoolVarP(&allowWrite, "write", "w", false, "Allow write access")
	pflag.StringVarP(&host, "host", "h", "[::]", "Host address to listen")
	pflag.IntVarP(&port, "port", "p", 8080, "Port to listen")
	pflag.StringVarP(&workingDir, "dir", "d", ".", "Directory to serve")
	pflag.StringVar(&tlsCert, "tsl-cert", "", "Path to an SSL/TLS certificate to serve with HTTPS")
	pflag.StringVar(&tlsKey, "tsl-key", "", "Path to an SSL/TLS certificate's private key")
	pflag.BoolVarP(&index, "index", "i", false, "Render index.html")

	if tlsCert == "" && tlsKey != "" {
		log.Fatalln("Missing SSL/TLS certificate's private key")
	}
	if tlsCert != "" && tlsKey == "" {
		log.Fatalln("Missing SSL/TLS certificate")
	}

	pflag.Parse()

	os.Chdir(workingDir)

	tmpl = make(map[string]*template.Template)
	tmpl["delete"] = template.Must(template.New("delete").Parse(htmlLayout + htmlDelete))
	tmpl["new-folder"] = template.Must(template.New("new-folder").Parse(htmlLayout + htmlNewFolder))
	tmpl["archive"] = template.Must(template.New("archive").Parse(htmlLayout + htmlArchive))
	tmpl["files"] = template.Must(template.New("files").Parse(htmlLayout + htmlFiles))
	tmpl["upload"] = template.Must(template.New("upload").Parse(htmlLayout + htmlUpload))
	tmpl["rename"] = template.Must(template.New("rename").Parse(htmlLayout + htmlRename))
	tmpl["edit"] = template.Must(template.New("edit").Parse(htmlLayout + htmlEdit))
	tmpl["search"] = template.Must(template.New("search").Parse(htmlLayout + htmlSearch))

	if !index {
		http.HandleFunc("GET /{path...}", servePath)
		http.HandleFunc("POST /action", action)
		if allowWrite {
			http.HandleFunc("POST /delete", delete)
			http.HandleFunc("POST /new-folder", newFolder)
			http.HandleFunc("POST /archive", archive)
			http.HandleFunc("POST /rename", rename)
			http.HandleFunc("POST /upload", upload)
			http.HandleFunc("POST /edit", edit)
		}
		http.HandleFunc("GET /__gofs__/style.css", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/css")
			io.WriteString(w, cssStyle)
		})
		http.HandleFunc("GET /__gofs__/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Write(faviconImg)
		})
		http.HandleFunc("GET /__gofs__/script.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/javascript")
			io.WriteString(w, scriptSource)
		})
	} else {
		_, err := os.Stat("index.html")
		if os.IsNotExist(err) {
			fmt.Println("Error: Cannot find index.html in the current folder.")
			return
		} else {
			fs := http.FileServer(http.Dir("."))
			http.Handle("/", fs)
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	var globalIp string
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.IsGlobalUnicast() {
			globalIp = ip.IP.String()
		}
	}

	var schema string
	if tlsCert != "" && tlsKey != "" {
		schema = "https"
	} else {
		schema = "http"
	}
	if host == "[::]" {
		fmt.Printf("Listening on: %v://%v:%v\n", schema, "localhost", port)
		fmt.Printf("              %v://%v:%v\n", schema, globalIp, port)
		fmt.Printf("              %v://%v:%v\n", schema, "[::1]", port)
	} else {
		fmt.Printf("Listening on: %v://%v:%v\n", schema, host, port)
	}
	if tlsCert == "" && tlsKey == "" {
		log.Fatal(http.ListenAndServe(fmt.Sprintf("%v:%v", host, port), nil))
	} else {
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("%v:%v", host, port), tlsCert, tlsKey, nil))
	}
}

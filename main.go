


package main


import "fmt"
import "os"
import "net/http"
import "io/ioutil"
import "strings"
import "strconv"
import "sort"

type Link struct {
    real		string
    replacing_name	string
}
type Links []Link
func (l Links) Len() int { return len(l) }
func (l Links) Less(i,j int) bool { return len([]rune(l[i].real)) > len([]rune(l[j].real)) }
func (l Links) Swap(i,j int) { l[i], l[j] = l[j], l[i] }
func (l *Links) Add(link *Link) {
    *l = append(*l, *link)
}
func (l *Links) ToMap(m map[string]string) {
    for _, v := range *l {
	m[v.real] = v.replacing_name
    }
}

var (
    main_url = get_site()
    queue_urls_for_saving = make(map[string]string)
    queue_stores_pages = make(map[string]string)
    page_index = 1
    chan_page_index = make(chan struct{})
)


func is_url_handled(u string) bool { // потоко опасно
    _, ok := queue_urls_for_saving[u]
    if ok { return true }
    
    _, ok = queue_stores_pages[u]
    if ok { return true }
    
    return false
}


func get_content_of_url(u string) (content string) {
    r, e := http.Get(u)
    if e != nil {
	content = e.Error()
	return
    }
    defer r.Body.Close()
    b, e := ioutil.ReadAll(r.Body)
    if e != nil {
	content = e.Error()
	return
    }
    content = string(b)
    return 
}


func get_links_in_content_and_transform_content(str string) (links map[string]string, content string) {
    links = make(map[string]string)
    content = str
    
    for {
	b := strings.Index(str, "<")
	e := strings.Index(str, ">")
	
	if b < 0 || e < 0 { break }
	
	sub := str[b:e + 1]
	href := strings.Index(sub, "href=")
	src := strings.Index(sub, "src=")
	
	if href > 0 {
	    end := ".html"
	    if strings.Index(sub, "link") > -1 { 
		if strings.Index(sub, "css") > -1 {
		    end = ".css"
		} else if strings.Index(sub, "xml") > -1 {
		    end = ".xml"
		}
	    }
	    
	    sub = sub[href + (len([]rune("href="))) + 1:]
	    sub = sub[:strings.Index(sub, "\"")]
	    
	    is_save := true
	    if end == ".html" {
		if sub[:1] == "/" { 
		    is_save = false
		} else if strings.Index(sub, "http") > -1 {
		    if strings.Index(sub, main_url) < 0 {
			is_save = false
		    }
		}
	    }
	    
	    if is_save {
		if sub[:1] == "/" {
		    sub = ".." + sub
		}
		links[sub] = strconv.Itoa(page_index) + end 
	    }
	} else if src > 0 {
	    sub = sub[src + (len([]rune("src="))) + 1:]
	    sub = sub[:strings.Index(sub, "\"")]
	    points := strings.Split(sub, ".")
	    end := "." + points[len(points) - 1]
	    if sub[:1] == "/" {
		sub = ".." + sub
	    }
	    links[sub] = strconv.Itoa(page_index) + end
	}
	str = str[e + 1:]
    }
    
    fmt.Println("-----------------------")
    slinks := Links{}
    for k, v := range links {
	slinks.Add(&Link{real: k, replacing_name: v })
    }
    
    sort.Sort(slinks)
    for _, v := range slinks {
	content = strings.ReplaceAll(content, v.real, v.replacing_name)
    }
    
    slinks.ToMap(links)
    
    return
}


func main() {
    content := get_content_of_url(main_url)
    links, content := get_links_in_content_and_transform_content(content)
    
    /*for k, v := range links {
	fmt.Println(k + " -> " + v)
    }*/
    //fmt.Println(content)
    
    fmt.Println("exit")
}

func get_site() string {
    return os.Args[1]
}
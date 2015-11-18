package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

//Template Html

var htmlTemplate string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Movies</title>
</head>
<body>
    <table border="1px">
        <thead>
            <tr>
                <th>Title</th>
                <th>Year</th>
            </tr>
        </thead>
        <tbody  id="cuerpo_tabla">
            {{range $movie := .movies}}
            <tr >
              <td>{{ $movie.Title }}</td>
              <td>{{ $movie.Year }}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    <button OnClick="Save();" style="margin: 2em 0 0 2em">Save Test Movie</button>
    <p>* Pls see java console</p>
</body>
<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.3/jquery.min.js"></script>
<script>
    String.format = function() {
        var theString = arguments[0];
        for (var i = 1; i < arguments.length; i++) {
            var regEx = new RegExp("\\{" + (i - 1) + "\\}", "gm");
            theString = theString.replace(regEx, arguments[i]);
        }
        return theString;
    };

    function getCookie(name) {
        var cookieValue = null;
        if (document.cookie && document.cookie != '') {
            var cookies = document.cookie.split(';');
            for (var i = 0; i < cookies.length; i++) {
                var cookie = jQuery.trim(cookies[i]);
                // Does this cookie string begin with the name we want?
                if (cookie.substring(0, name.length + 1) == (name + '=')) {
                    cookieValue = decodeURIComponent(cookie.substring(name.length + 1));
                    break;
                }
            }
        }
        return cookieValue;
    }

	// FIXME: I try two methods to send csrftoken, but don't work :'(

    //var csrftoken = getCookie('csrftoken');
    var csrftoken = getCookie('{{.CsrfToken}}');



    $.ajaxSetup({
        contentType: 'application/json',
        dataType: "JSON",
        processData: false,
        beforeSend: function (xhr) {
            xhr.setRequestHeader("X-CSRF-Token", csrftoken);
        }
    });

    function Save() {
        var url = "/api/movie/testimdbKey";
        var request = 'POST';
        var objetos = {"title": "Avatar",
                       "rating": "9.9",
                       "year": "2012"};

        $.ajax({url: url,
            type: request,
            async: false,
            data: JSON.stringify(objetos)
        }).error(function (re){
            console.log("oh cmon!");
            console.log(re);
        })
        .success(function (rs) {
            console.log("oh yeah!");
            console.log(rs);
        });
    }

    $(document).ready(function(){
        var plantilla = '<tr> \
                            <td>{0}</td> \
                            <td>{1}</td> \
                        </tr>';
        $.ajax({
            url: '/api/movie',
            type: 'GET'
        }).complete(function(c){
                    console.log("GET works!");
                    console.log(c);
                    if (c.status == 302){
                        var moviesStr;
                        var movieArray = c.responseJSON;
                        for (var i = 0; i < movieArray.length; i++) {
                            moviesStr = String.format(plantilla, movieArray[i].title,
                                                                 movieArray[i].year);
                            $("#cuerpo_tabla").append(moviesStr);
                        }
                    }
                });
    });
</script>
</html>`

// Movie Struct
type Movie struct {
	Title  string `json:"title"`
	Rating string `json:"rating"`
	Year   string `json:"year"`
}

// Movie Database
var movies = map[string]*Movie{
	"tt0076759": &Movie{Title: "Star Wars: A New Hope", Rating: "8.7", Year: "1977"},
	"tt0082971": &Movie{Title: "Indiana Jones: Raiders of the Lost Ark", Rating: "8.6", Year: "1981"},
}

func main() {

	muxRouter := mux.NewRouter()

	muxRouter.HandleFunc("/", getMovies).Methods("GET")
	//-------Api ----------
	muxRouter.HandleFunc("/api/movie", apiMovies).Methods("GET")
	muxRouter.HandleFunc("/api/movie/{imdbKey}", apiMovies).Methods("GET", "POST", "DELETE", "PUT")

	CSRF := csrf.Protect([]byte("_yJ3~mY#i{l}=mt^kIUfxW9o:Z~ydF(."))

	go http.ListenAndServe(":3000", CSRF(muxRouter))

	log.Println("Servidor Iniciado http://127.0.0.1:3000/, presiona <ENTER> para salir.")
	fmt.Scanln()

}

// ---Show Movies and ajax operation
func getMovies(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content Type", "text/html")

	t, _ := template.New("html").Parse(htmlTemplate)
	params := map[string]interface{}{"movies": movies,
		                             "CsrfToken": csrf.Token(req),}
	t.Execute(w, params)

}

// -- Api handler
func apiMovies(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(req)
	imdbKey := vars["imdbKey"]

	switch req.Method {

	case "GET":
		var movie map[string]*Movie

		//Get all Movies
		if imdbKey == "" {
			movie = movies
		} else {
			//Get if exist
			movieReturn, ok := movies[imdbKey]
			if ok {
				movie = map[string]*Movie{imdbKey: movieReturn}
			}
		}

		if len(movie) == 0 {
			outgoingJSON, _ := json.Marshal(map[string]string{"error": "Movie not found"})
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(outgoingJSON))
			return
		}

		outgoingJSON, err := json.Marshal(movie)
		if err != nil {
			ErrorControl(w, err)
			return
		}
		w.WriteHeader(http.StatusFound)
		w.Write([]byte(outgoingJSON))

	case "DELETE":
		delete(movies, imdbKey)
		w.WriteHeader(http.StatusNoContent)
	case "POST":
		movie := new(Movie)
		decoder := json.NewDecoder(req.Body)
		err_d := decoder.Decode(&movie)

		if err_d != nil {
			ErrorControl(w, err_d)
			return
		}

		movies[imdbKey] = movie
		outgoingJSON, err_m := json.Marshal(movie)

		if err_m != nil {
			ErrorControl(w, err_d)
			return
		}
		w.Write([]byte(outgoingJSON))
	}
}

func ErrorControl(w http.ResponseWriter, err_r error) {
	http.Error(w, err_r.Error(), http.StatusInternalServerError)
}

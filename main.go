package main

import (
	"math/rand"
	"net/http"
	"time"
)

var images = []string{
	"https://randomon.crstian.me/chibimon_angry.gif",
	"https://media4.giphy.com/media/v1.Y2lkPTc5MGI3NjExanZzY3RjdHdmNjNyMzZjOWd5bmRiYWdrc2swM2tpbmt0dWIzank1aiZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/lEc7pLgBpipbi/giphy.webp",
	"https://media0.giphy.com/media/v1.Y2lkPTc5MGI3NjExbHk1d205OWI2aWg2ejc4NGt0d3A0YnZ4bzh2cXRja3BqdG9tZncxaCZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/qgkrtsDy4MhLq/giphy.webp",
	"https://media2.giphy.com/media/slVWEctHZKvWU/giphy.gif?cid=ecf05e476ptp3hv2w3xzo3pbiupfyinr0o0u1jmzjhshvq89&rid=giphy.gif&ct=g",
	// Añade todas las URLs de los GIFs y WebP que subiste
}

func randomImageHandler(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(images))
	http.Redirect(w, r, images[randomIndex], http.StatusFound)
}

func main() {
	http.HandleFunc("/", randomImageHandler)
	// En un entorno no serverless, usaríamos:
	http.ListenAndServe(":3000", nil)
}

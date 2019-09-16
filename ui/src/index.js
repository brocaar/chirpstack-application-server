import React from "react";
import ReactDOM from "react-dom";

import "typeface-roboto";
import Leaflet from "leaflet";

import App from "./App";

import "leaflet/dist/leaflet.css";
import "leaflet.awesome-markers/dist/leaflet.awesome-markers.css";
import "codemirror/lib/codemirror.css";
import "codemirror/theme/base16-light.css";
import "react-leaflet-markercluster/dist/styles.min.css";
import "@fortawesome/fontawesome-free/css/all.min.css";
import "./index.css";

Leaflet.Icon.Default.imagePath = "//cdnjs.cloudflare.com/ajax/libs/leaflet/1.0.0/images/"

ReactDOM.render(<App />, document.getElementById("root"));

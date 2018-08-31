import React from "react";
import ReactDOM from "react-dom";

import "typeface-roboto";
import Leaflet from "leaflet";

import App from "./App";

import "leaflet/dist/leaflet.css";
import "codemirror/lib/codemirror.css";
import "codemirror/theme/base16-light.css";
import "./index.css";

Leaflet.Icon.Default.imagePath = "//cdnjs.cloudflare.com/ajax/libs/leaflet/1.0.0/images/"

ReactDOM.render(<App />, document.getElementById("root"));

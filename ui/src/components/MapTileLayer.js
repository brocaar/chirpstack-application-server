import React, { Component } from 'react';

import { TileLayer } from 'react-leaflet';


class MapTileLayer extends Component {
  render() {
    return(
      <TileLayer
        url='//{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
        attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
      />
    )
  }
}

export default MapTileLayer;

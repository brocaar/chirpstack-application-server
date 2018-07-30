import React, { Component } from "react";

import JSONTreeOriginal from "react-json-tree";


class JSONTree extends Component {
  render() {
    const theme = {
      scheme: 'google',
      author: 'seth wright (http://sethawright.com)',
      base00: '#000000',
      base01: '#282a2e',
      base02: '#373b41',
      base03: '#969896',
      base04: '#b4b7b4',
      base05: '#c5c8c6',
      base06: '#e0e0e0',
      base07: '#ffffff',
      base08: '#CC342B',
      base09: '#F96A38',
      base0A: '#FBA922',
      base0B: '#198844',
      base0C: '#3971ED',
      base0D: '#3971ED',
      base0E: '#A36AC7',
      base0F: '#3971ED',
    }

    return(
      <JSONTreeOriginal
        data={this.props.data}
        theme={theme}
        hideRoot={true}
        shouldExpandNode={() => {return true}}
      />
    );
  }
}

export default JSONTree;

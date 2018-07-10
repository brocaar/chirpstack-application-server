import React, { Component } from 'react';


class Loaded extends Component {
  constructor() {
    super();
    
    this.state = {
      loaded: false,
    };
    
    this.testLoaded = this.testLoaded.bind(this);
  }
  
  componentDidMount() {
    this.testLoaded(this.props.loaded);
  }
  
  componentWillReceiveProps(newProps) {
    this.testLoaded(newProps.loaded);
  }
  
  testLoaded(obj) {
    for (const key of Object.keys(obj)) {
      var loaded = true;
      
      if (obj[key] === false) {
        loaded = false;
      }
    }
    
    this.setState({
      loaded: loaded,
    });
  }
  
  render() {
    return(
      <div>
        {this.state.loaded && this.props.children}
      </div>
    );
  }
}

export default Loaded;

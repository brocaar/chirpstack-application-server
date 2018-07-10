import { Component } from "react";


class FormComponent extends Component {
  constructor() {
    super();

    this.state = {};

    this.onChange = this.onChange.bind(this);
    this.onSubmit = this.onSubmit.bind(this);
  }

  onChange(e) {
    let lookup = e.target.id.split(".");
    const field = lookup[lookup.length-1];
    lookup.pop(); // remove last item

    let object = this.state.object;
    let obj = object;
    for (const f of lookup) {
      obj = obj[f];
    }

    if (e.target.type === "checkbox") {
      obj[field] = e.target.checked;
    } else if (e.target.type === "number") {
      obj[field] = parseInt(e.target.value, 10);
    } else {
      obj[field] = e.target.value;
    }

    this.setState({
      object: object,
    })
  }

  onSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.object);
  }

  componentDidMount() {
    this.setState({
      object: this.props.object || {},
    });
  }

  componentDidUpdate(prevProps) {
    if (prevProps.object !== this.props.object) {
      this.setState({
        object: this.props.object || {},
      });
    }
  }
}

export default FormComponent;

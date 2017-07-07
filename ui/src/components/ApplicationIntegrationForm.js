import React, { Component } from 'react';

import Select from "react-select";


class ApplicationHTTPIntegrationForm extends Component {
  constructor() {
    super();
    this.onChange = this.onChange.bind(this);
  }

  onChange(field, e) {
    let integration = this.props.integration;
    integration[field] = e.target.value;

    this.props.onFormChange(integration);
  }

  render() {
    return(
      <div>
        <div className="form-group">
          <label className="control-label" htmlFor="dataUpURL">Uplink data URL</label>
          <input className="form-control" id="dataUpURL" name="dataUpURL" type="text" placeholder="http://example.com/uplink" value={this.props.integration.dataUpURL || ''} onChange={this.onChange.bind(this, 'dataUpURL')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="joinNotificationURL">Join notification URL</label>
          <input className="form-control" id="joinNotificationURL" name="joinNotificationURL" type="text" placeholder="http://example.com/join" value={this.props.integration.joinNotificationURL || ''} onChange={this.onChange.bind(this, 'joinNotificationURL')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="ackNotificationURL">ACK notification URL</label>
          <input className="form-control" id="ackNotificationURL" name="ackNotificationURL" type="text" placeholder="http://example.com/ack" value={this.props.integration.ackNotificationURL || ''} onChange={this.onChange.bind(this, 'ackNotificationURL')} />
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="errorNotificationURL">Error notification URL</label>
          <input className="form-control" id="errorNotificationURL" name="errorNotificationURL" type="text" placeholder="http://example.com/error" value={this.props.integration.errorNotificationURL || ''} onChange={this.onChange.bind(this, 'errorNotificationURL')} />
        </div>
      </div>
    );
  }
}

class ApplicationIntegrationForm extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      integration: {},
      kindDisabled: false,
    };

    this.handleSubmit = this.handleSubmit.bind(this);
    this.onKindSelect = this.onKindSelect.bind(this);
    this.onFormChange = this.onFormChange.bind(this);
  }

  componentDidMount() {
    this.setState({
      integration: this.props.integration,
    });

    if (typeof(this.props.integration.kind) !== "undefined") {
      this.setState({
        kindDisabled: true,
      });
    }
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      integration: nextProps.integration,
    });

    if (nextProps.integration.kind !== "") {
      this.setState({
        kindDisabled: true,
      });
    }
  }

  onChange(field, e) {
    let integration = this.state.integration;
    integration[field] = e.target.value;

    this.setState({
      integration: integration,
    });
  }

  onFormChange(integration) {
    this.setState({
      integration: integration,
    });
  }

  onKindSelect(val) {
    let integration = this.state.integration;
    integration.kind = val.value;
    this.setState({
      integration: integration,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.integration);
  }

  render() {
    const kindOptions = [
      {value: "http", label: "HTTP integration"},
    ];

    let form = <div></div>;

    if (this.state.integration.kind === "http") {
      form = <ApplicationHTTPIntegrationForm integration={this.state.integration} onFormChange={this.onFormChange} />;
    }

    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="kind">Integration kind</label>
          <Select
            name="kind"
            value={this.state.integration.kind}
            options={kindOptions}
            onChange={this.onKindSelect}
            clearable={false}
            disabled={this.state.kindDisabled}
          />
        </div>
        {form}
        <div className="btn-toolbar pull-right">
          <a className="btn btn-default" onClick={this.context.router.goBack}>Go back</a>
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}

export default ApplicationIntegrationForm;

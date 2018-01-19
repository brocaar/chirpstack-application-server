import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import Select from "react-select";


class ApplicationHTTPIntegrationHeaderForm extends Component {
  constructor() {
    super();
    this.onChange = this.onChange.bind(this);
    this.onDelete = this.onDelete.bind(this);
  }

  onChange(field, e) {
    let header = this.props.header;
    header[field] = e.target.value;

    this.props.onHeaderChange(this.props.index, header);
  }

  onDelete(e) {
    this.props.onDeleteHeader(this.props.index);
    e.preventDefault();
  }

  render() {
    return(
      <div className="form-group row">
        <div className="col-sm-4">
          <input type="text" className="form-control" placeholder="Header name" value={this.props.header.key || ''} onChange={this.onChange.bind(this, 'key')} />
        </div>
        <div className="col-sm-7">
          <input type="text" className="form-control" placeholder="Header value" value={this.props.header.value || ''} onChange={this.onChange.bind(this, 'value')} />
        </div>
        <div className="col-sm-1">
          <button type="button" className="btn btn-link pull-right" onClick={this.onDelete}>
            <span className="glyphicon glyphicon-remove" aria-hidden="true"></span>
          </button>
        </div>
      </div>
    );
  }
}

class ApplicationHTTPIntegrationForm extends Component {
  constructor() {
    super();
    this.onChange = this.onChange.bind(this);
    this.onHeaderChange = this.onHeaderChange.bind(this);
    this.addHeader = this.addHeader.bind(this);
    this.onDeleteHeader = this.onDeleteHeader.bind(this);
  }

  onChange(field, e) {
    let integration = this.props.integration;
    integration[field] = e.target.value;

    this.props.onFormChange(integration);
  }

  addHeader(e) {
    let integration = this.props.integration;
    if (typeof(integration.headers) === "undefined") {
      integration.headers = [{}];
    } else {
      integration.headers.push({});
    }

    this.props.onFormChange(integration);

    e.preventDefault();
  }

  onHeaderChange(index, header) {
    let integration = this.props.integration;
    integration.headers[index] = header;

    this.props.onFormChange(integration);
  }

  onDeleteHeader(index) {
    let integration = this.props.integration;
    integration.headers.splice(index, 1);

    this.props.onFormChange(integration);
  }

  render() {
    let headers = [];
    if (typeof(this.props.integration.headers) !== "undefined") {
      headers = this.props.integration.headers;
    }

    const HTTPHeaders = headers.map((header, i) => <ApplicationHTTPIntegrationHeaderForm key={i} index={i} header={header} onHeaderChange={this.onHeaderChange} onDeleteHeader={this.onDeleteHeader} />);

    return(
      <div>
        <fieldset>
          <legend>Headers</legend>
          {HTTPHeaders}
          <div className="form-group">
            <button className="btn btn-default pull-right" onClick={this.addHeader}>Add header</button>
          </div>
        </fieldset>
        <fieldset>
          <legend>Endpoints</legend>
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
        </fieldset>
      </div>
    );
  }
}

class ApplicationIntegrationForm extends Component {
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
          <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}

export default withRouter(ApplicationIntegrationForm);

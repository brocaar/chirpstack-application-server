import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';

import SessionStore from "../stores/SessionStore";


class OrganizationForm extends Component {
  constructor() {
    super();

    this.state = {
      organization: {},
      showCanHaveGateways: SessionStore.isAdmin(),
    };

    this.handleSubmit = this.handleSubmit.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      organization: nextProps.organization,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.props.onSubmit(this.state.organization);
  }

  onChange(field, e) {
    let organization = this.state.organization;
    if (e.target.type === "checkbox") {
      organization[field] = e.target.checked; 
    } else {
      organization[field] = e.target.value;
    }
    this.setState({
      organization: organization,
    });
  }

  render() {
    return(
      <form onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Organization name</label>
          <input className="form-control" id="name" type="text" placeholder="e.g. my-organization" pattern="[\w-]+" required value={this.state.organization.name || ''} onChange={this.onChange.bind(this, 'name')} />
          <p className="help-block">
            The name may only contain words, numbers and dashes. 
          </p>
        </div>
        <div className="form-group">
          <label className="control-label" htmlFor="name">Display name</label>
          <input className="form-control" id="name" type="text" placeholder="My Organization" required value={this.state.organization.displayName || ''} onChange={this.onChange.bind(this, 'displayName')} />
        </div>
        <div className={"form-group " + (this.state.showCanHaveGateways ? '' : 'hidden')}>
          <label className="control-label">Can have gateways</label>
          <div className="checkbox">
            <label>
              <input type="checkbox" name="canHaveGateways" id="canHaveGateways" checked={!!this.state.organization.canHaveGateways} onChange={this.onChange.bind(this, 'canHaveGateways')} /> Can have gateways 
            </label>
          </div>
          <p className="help-block">
            When checked, it means that organization administrators are able to add their own gateways to the network. 
            Note that the usage of the gateways is not limited to this organization.
          </p>
        </div>
        <hr />
        <div className="btn-toolbar pull-right">
          <a className="btn btn-default" onClick={this.props.history.goBack}>Go back</a>
          <button type="submit" className="btn btn-primary">Submit</button>
        </div>
      </form>
    );
  }
}

export default withRouter(OrganizationForm)

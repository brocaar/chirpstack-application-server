import React, { Component } from 'react';
import { Link } from 'react-router';

import Select from "react-select";

import OrganizationStore from "../stores/OrganizationStore";
import SessionStore from "../stores/SessionStore";


class OrganizationSelect extends Component {
  static contextTypes = {
    router: React.PropTypes.object.isRequired
  };

  constructor() {
    super();

    this.state = {
      organization: {},
      showDropdown: false,
      initialOptions: [],
    };

    this.setSelectedOrganization = this.setSelectedOrganization.bind(this);
    this.setInitialOrganizations = this.setInitialOrganizations.bind(this);
    this.onSelect = this.onSelect.bind(this);
  }

  componentWillMount() {
    OrganizationStore.getOrganization(this.props.organizationID, (org) => {
      this.setState({
        organization: org,
      }); 
      this.setSelectedOrganization();
      this.setOrganizationID(this.props.organizationID);
    });

    OrganizationStore.getAll("", 2, 0, (totalCount, orgs) => {
      if (totalCount > 1) {
        this.setState({
          showDropdown: true,
        });
      }
    });
  }

  componentWillReceiveProps(nextProps) {
    if (nextProps.organizationID !== this.state.organization.id) {
      OrganizationStore.getOrganization(nextProps.organizationID, (org) => {
        this.setState({
          organization: org,
        });
        this.setSelectedOrganization();
      });
    }
  }

  setSelectedOrganization() {
    this.setState({
      initialOptions: [{
        value: this.state.organization.id,
        label: this.state.organization.displayName,
      }],
    }); 
  }

  setInitialOrganizations() {
    OrganizationStore.getAll("", 10, 0, (totalCount, orgs) => {
      const options = orgs.map((org, i) => {
        return {
          value: org.id,
          label: org.displayName,
        };
      });

      this.setState({
        initialOptions: options,
      });
    });  
  }

  onSelect(val) {
    SessionStore.setOrganizationID(val.value);
    this.context.router.push('/organizations/'+val.value);
  }

  onAutocomplete(input, callbackFunc) {
    OrganizationStore.getAll(input, 10, 0, (totalCount, orgs) => {
      const options = orgs.map((org, i) => {
        return {
          value: org.id,
          label: org.displayName,
        };
      });

      callbackFunc(null, {
        options: options,
        complete: true,
      });
    });
  }

  render() {
    let org;

    if (this.state.showDropdown) {
      org = <div className="org-select"><Select
        name="organization"
        options={this.state.initialOptions}
        value={this.props.organizationID}
        clearable={false}
        autosize={true}
        onOpen={this.setInitialOrganizations}
        onClose={this.setSelectedOrganization}
        loadOptions={this.onAutocomplete}
        autoload={false}
        onChange={this.onSelect}
      /></div>
    } else {
      org = <Link to={`/organizations/${this.state.organization.id}`}>{this.state.organization.displayName}</Link>;
    }

    return(org);
  }
}

export default OrganizationSelect;

import React, { Component } from "react";
import { Route, Switch, Link, withRouter } from 'react-router-dom';

import SessionStore from "../../stores/SessionStore";
import ApplicationStore from "../../stores/ApplicationStore";
import OrganizationSelect from "../../components/OrganizationSelect";

// devices
import ListNodes from '../nodes/ListNodes';
import CreateNode from "../nodes/CreateNode";

// applications
import UpdateApplication from "./UpdateApplication";
import ApplicationIntegrations from "./ApplicationIntegrations";
import CreateApplicationIntegration from "./CreateApplicationIntegration";
import UpdateApplicationIntegration from "./UpdateApplicationIntegration";


class ApplicationLayout extends Component {
  constructor() {  
    super();

    this.state = {
      application: {},
      isAdmin: false,
    };

    this.onDelete = this.onDelete.bind(this);
  }

  componentDidMount() {
    ApplicationStore.getApplication(this.props.match.params.applicationID, (application) => {
      this.setState({application: application});
    });

    this.setState({
      isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
    });

    SessionStore.on("change", () => {
      this.setState({
        isAdmin: (SessionStore.isAdmin() || SessionStore.isOrganizationAdmin(this.props.match.params.organizationID)),
      });
    });
  }

  onDelete() {
    if (window.confirm("Are you sure you want to delete this application?")) {
      ApplicationStore.deleteApplication(this.props.match.params.applicationID, (responseData) => {
        this.props.history.push(`/organizations/${this.props.match.params.organizationID}/applications`);
      }); 
    }
  }

  render() {
    let activeTab = this.props.location.pathname.replace(this.props.match.url, '');

    return (
      <div>
        <ol className="breadcrumb">
          <li><Link to="/organizations">Organizations</Link></li>
          <li><OrganizationSelect organizationID={this.props.match.params.organizationID} /></li>
          <li><Link to={`/organizations/${this.props.match.params.organizationID}/applications`}>Applications</Link></li>
          <li className="active">{this.state.application.name}</li>
        </ol>
        <div className={(this.state.isAdmin ? '' : 'hidden')}>
          <div className="clearfix">
            <div className="btn-group pull-right" role="group" aria-label="...">
              <a><button type="button" className="btn btn-danger" onClick={this.onDelete}>Delete application</button></a>
            </div>
          </div>
        </div>
        <ul className="nav nav-tabs">
          <li role="presentation" className={(activeTab === "" || activeTab === "/nodes/create") ? 'active' : ''}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}`}>Devices</Link></li>
          <li role="presentation" className={(activeTab === "/edit" ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/edit`}>Application configuration</Link></li>
          <li role="presentation" className={((activeTab === "/integrations" || activeTab === "/integrations/create" || activeTab === "/integrations/http") ? 'active' : '') + (this.state.isAdmin ? '' : 'hidden')}><Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations`}>Integrations</Link></li>
        </ul>
        <hr />
        <Switch>
          <Route exact path={this.props.match.path} component={ListNodes} />
          <Route exact path={`${this.props.match.path}/edit`} component={UpdateApplication} />
          <Route exact path={`${this.props.match.path}/nodes/create`} component={CreateNode} />
          <Route exact path={`${this.props.match.path}/integrations`} component={ApplicationIntegrations} />
          <Route exact path={`${this.props.match.path}/integrations/create`} component={CreateApplicationIntegration} />
          <Route exact path={`${this.props.match.path}/integrations/http`} component={UpdateApplicationIntegration} />
        </Switch>
      </div>
    );
  }
}

export default withRouter(ApplicationLayout);

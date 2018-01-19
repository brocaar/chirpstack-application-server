import React, { Component } from 'react';
import { Link } from 'react-router-dom';

import ApplicationStore from "../../stores/ApplicationStore";

const integrationMap = {
  HTTP: {
    name: 'HTTP integration',
    endpoint: 'http',
  },
};


class IntegrationRow extends Component {
  render() {
    return(
      <tr>
        <td><Link to={`/organizations/${this.props.params.organizationID}/applications/${this.props.params.applicationID}/integrations/http`}>{integrationMap[this.props.kind].name}</Link></td>
      </tr>
    );
  }
}

class ApplicationIntegrations extends Component {
  constructor() {
    super();

    this.state = {
      integrations: [],
    };
  }

  componentDidMount() {
    ApplicationStore.listIntegrations(this.props.match.params.applicationID, (integrations) => {
      this.setState({
        integrations: integrations.kinds,
      });
    });    
  }

  render() {
    const IntegrationRows = this.state.integrations.map((integration, i) => <IntegrationRow key={integration} kind={integration} params={this.props.match.params} />);

    return(
      <div className="panel panel-default">
        <div className="panel-heading clearfix">
          <div className="btn-group pull-right">
           <Link to={`/organizations/${this.props.match.params.organizationID}/applications/${this.props.match.params.applicationID}/integrations/create`}><button type="button" className="btn btn-default btn-sm">Add integration</button></Link>
          </div>
        </div>
        <div className="panel-body">
          <table className="table table-hover">
            <thead>
              <tr>
                <th>Kind</th>
              </tr>
            </thead>
            <tbody>
              {IntegrationRows}
            </tbody>
          </table>
        </div>
      </div>
    );
  }
}

export default ApplicationIntegrations;

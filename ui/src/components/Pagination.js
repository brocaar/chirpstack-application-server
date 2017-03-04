import React, { Component } from 'react';
import { Link } from 'react-router';

class PaginationPage extends Component {
  render() {
    return(
      <li className={this.props.pageNumber === parseInt(this.props.currentPage, 10) ? 'active' : ''}><Link to={{pathname: this.props.pathname, query: {page: this.props.pageNumber}}}>{this.props.pageNumber}</Link></li>
    );
  }
}

class Pagination extends Component {
  render() {
    let pages = [];
    for (let i = 0; i < this.props.pages; i++) {
      pages.push(<PaginationPage key={i} pageNumber={i+1} currentPage={this.props.currentPage} pathname={this.props.pathname} />);
    }

    return(
      <nav className={this.props.pages === 1 ? 'hidden' : 'pull-right'}>
        <ul className="pagination">
          {pages}
        </ul>
      </nav> 
    );
  }
}

export default Pagination;

import React, { Component } from 'react';
import { Link } from 'react-router-dom';

class PaginationPage extends Component {
  render() {
    return(
      <li className={this.props.pageNumber === parseInt(this.props.currentPage, 10) ? 'active' : ''}><Link to={{pathname: this.props.pathname, search: `?page=${this.props.pageNumber}`}}>{this.props.pageNumber}</Link></li>
    );
  }
}

class SkipPagination extends Component {
  render() {
    return(
      <li className=""><Link to={{pathname: this.props.pathname, search: `?page=${this.props.pageNumber}`}}>...</Link></li>
    );
  }
}

class Pagination extends Component {
  render() {
    let startPages = [];
    let middlePages = [];
    let endPages = [];
    let startSkip = "";
    let endSkip = "";
    let pageSteps = 3;

    //Check if pageSteps are set in props
    if(this.props.pageSteps && parseInt(this.props.pageSteps, 10) > 0) {
      pageSteps = this.props.pageSteps;
    }

    let currentPage = parseInt(this.props.currentPage, 10);
    let totalPages = parseInt(this.props.pages, 10);

    //If all needed steps, along with first, last and current page may be rendered, there's no need to cut pagination.
    if(totalPages < pageSteps * 2 + 3) {
      for(let i = 0; i < totalPages; i++) {
        middlePages.push(<PaginationPage key={i} pageNumber={i+1} currentPage={this.props.currentPage} pathname={this.props.pathname} />);
      }
    }
    else {
      //Check if active page belongs to start, middle or end.
      if(currentPage <= pageSteps + 2) {
        endPages.push(<PaginationPage key={totalPages-1} pageNumber={totalPages} currentPage={this.props.currentPage} pathname={this.props.pathname} />)
        for(let i = 0; i < currentPage + pageSteps; i++) {
          startPages.push(<PaginationPage key={i} pageNumber={i+1} currentPage={this.props.currentPage} pathname={this.props.pathname} />)
        }
        let endSkipPage = totalPages - Math.ceil((totalPages - (currentPage + pageSteps)) / 2);
        endSkip = <SkipPagination key={endSkipPage-1} pageNumber={endSkipPage} pathname={this.props.pathname} />
      } else if (currentPage >= totalPages - (pageSteps + 2)) {
        startPages.push(<PaginationPage key={0} pageNumber={1} currentPage={this.props.currentPage} pathname={this.props.pathname} />)
        for(let i = currentPage - pageSteps - 1; i < totalPages; i++) {
          endPages.push(<PaginationPage key={i} pageNumber={i+1} currentPage={this.props.currentPage} pathname={this.props.pathname} />)
        }
        let startSkipPage = Math.ceil((currentPage - pageSteps)/2);
        startSkip = <SkipPagination key={startSkipPage-1} pageNumber={startSkipPage} pathname={this.props.pathname} />
      } else {
        startPages.push(<PaginationPage key={0} pageNumber={1} currentPage={this.props.currentPage} pathname={this.props.pathname} />)
        endPages.push(<PaginationPage key={totalPages-1} pageNumber={totalPages} currentPage={this.props.currentPage} pathname={this.props.pathname} />)
        for(let i = currentPage - pageSteps - 1; i < currentPage + pageSteps; i++) {
          middlePages.push(<PaginationPage key={i} pageNumber={i+1} currentPage={this.props.currentPage} pathname={this.props.pathname} />)
        }
        let startSkipPage = Math.ceil((currentPage - pageSteps)/2);
        startSkip = <SkipPagination key={startSkipPage-1} pageNumber={startSkipPage} pathname={this.props.pathname} />
        let endSkipPage = totalPages - Math.ceil((totalPages - (currentPage + pageSteps)) / 2);
        endSkip = <SkipPagination key={endSkipPage-1} pageNumber={endSkipPage} pathname={this.props.pathname} />
      }
    }

    return(
      <nav className={this.props.pages === 1 ? 'hidden' : 'pull-right'}>
        <ul className="pagination">
          {startPages}{startSkip}{middlePages}{endSkip}{endPages}
        </ul>
      </nav>
    );
  }
}

export default Pagination;

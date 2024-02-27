import React from 'react';
import { useNavigate } from 'react-router-dom';
import PaginatedTable from './PaginatedTable';

const Home = (props) => {
  const { fixture } = props;

  return (
    <div className="mainContainer">
      <PaginatedTable pagination={fixture} />{' '}
    </div>
  );
};

export default Home;

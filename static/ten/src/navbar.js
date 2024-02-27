import React from 'react';
import Button from '@mui/material/Button';

import { useNavigate } from 'react-router-dom';

function Navbar(props) {
  const { loggedIn, setLoggedIn } = props;
  const navigate = useNavigate();

  const onButtonClick = () => {
    if (loggedIn) {
      setLoggedIn(false);
    } else {
      navigate('/login');
    }
  };

  return (
    <nav>
      <a href="/">
        <span className="logo"> </span>{' '}
      </a>{' '}
      <Button variant="contained" onClick={onButtonClick}>
        {' '}
        {loggedIn ? 'Log out' : 'Log in'}{' '}
      </Button>{' '}
    </nav>
  );
}

export default Navbar;

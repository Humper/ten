import {
  Admin,
  Resource,
  ListGuesser,
  EditGuesser,
  ShowGuesser,
} from 'react-admin';
import { IpList } from './IPList';
import { UserList } from './UserList';
import { UserEdit } from './UserEdit';
import { UserCreate } from './UserCreate';
import './App.css';
import dataProvider from './dataProvider';
import authProvider from './authProvider';

const App = () => (
  <Admin dataProvider={dataProvider} authProvider={authProvider}>
    <Resource name="IPs" list={IpList} />{' '}
    <Resource name="users" list={UserList} edit={UserEdit} show={ShowGuesser} create={UserCreate} />{' '}
  </Admin>
);

export default App;

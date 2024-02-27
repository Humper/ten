const apiUrl = 'http://greg.isthebest.com:8080';

const authProvider = {
  login: ({ username, password }) => {
    const request = new Request(apiUrl + '/login', {
      method: 'POST',
      body: JSON.stringify({
        email: username,
        password,
      }),
      headers: new Headers({ 'Content-Type': 'application/json' }),
      credentials: 'include',
    });
    return fetch(request)
      .then((response) => {
        if (response.status < 200 || response.status >= 300) {
          throw new Error(response.statusText);
        }
        return response.json();
      })
      .then((auth) => {
        localStorage.setItem('user', JSON.stringify(auth));
      })
      .catch(() => {
        throw new Error('Network error');
      });
  },
  checkAuth: () => {
    return localStorage.getItem('user') ? Promise.resolve() : Promise.reject();
  },
  getPermissions: () => {
    // Required for the authentication to work
    return Promise.resolve();
  },
  checkError: (error) => {
    // Required for the authentication to work
    return Promise.resolve();
  },
  getIdentity: () => {
    try {
      const { ID, Name } = JSON.parse(localStorage.getItem('user'));
      return Promise.resolve({ id: ID, fullName: Name});
    } catch (error) {
      return Promise.reject(error);
    }
  },
  logout: (error) => {
    localStorage.removeItem('user');
    return Promise.resolve();
  },
  // ...
};

export default authProvider;

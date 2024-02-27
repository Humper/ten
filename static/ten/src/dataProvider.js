import { DataProvider, fetchUtils } from 'react-admin';
import { stringify } from 'query-string';

const apiUrl = 'http://localhost:8080';
const httpClient = (url, options={}) => {
  if (!options.headers) {
    options.headers = new Headers({ Accept: 'application/json' });
  }
  options.credentials = 'include';
  return fetchUtils.fetchJson(url, options);
}

const resourceMap = {
  IPs: 'tor',
  users: 'users',
  country_code: 'country_code',
};

const dataProvider = {
  getList: (resource, params) => {
    const { page, perPage } = params.pagination;
    const { field, order } = params.sort;
    const { filter } = params;

    const query = {
      sort: field + ' ' + order,
      page: JSON.stringify(page),
      limit: JSON.stringify(perPage),
      filter: JSON.stringify(filter),
    };

    const url = `${apiUrl}/${resourceMap[resource]}?${stringify(query)}`;

    return httpClient(url).then(({ headers, json }) => ({
      data: json.rows.map((row) => ({ ...row, id: row.ID })),
      total: json.total_rows,
    }));
  },
  getOne: (resource, params) => {
    const { id } = params;

    const url = `${apiUrl}/${resourceMap[resource]}/${id}`;
    return httpClient(url).then(({ headers, json }) => {
      json.id = json.ID;
      return {
      data: json,
    }
  });
  },
  update: (resource, params) => {
    const { id, data } = params;
    const url = `${apiUrl}/${resourceMap[resource]}/${id}`;

    if (data.AllowedIPs) {
      data.AllowedIPs = data.AllowedIPs.split(',');
    }
    return httpClient(url, {
      method: 'PUT',
      body: JSON.stringify(data),
    }).then(({ json }) => {
      json.id = json.ID;
      return { data: json };  
    });
  },  
  delete: (resource, params) => {
    const { id } = params;
    const url = `${apiUrl}/${resourceMap[resource]}/${id}`;

    return httpClient(url, {
      method: 'DELETE',
    }).then(({ json }) => {
      json.id = json.ID;
      return { data: json };  
    });
  },  
  create: (resource, params) => {
    const { data } = params;
    const url = `${apiUrl}/${resourceMap[resource]}`;

    if (data.AllowedIPs) {
      data.AllowedIPs = data.AllowedIPs.split(',');
    }
    return httpClient(url, {
      method: 'POST',
      body: JSON.stringify(data),
    }).then(({ json }) => {
      json.id = json.ID;
      return { data: json };  
    });
  },
};

export default dataProvider;

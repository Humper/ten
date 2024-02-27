import React from 'react';

const PaginatedTable = (props) => {
  const { pagination } = props;
  const { limit, page, sort, total_rows, total_pages, rows } = pagination;

  return (
    <div className="tableContainer">
      <table>
        <thead>
          <tr>
            <th> IP </th> <th> CountryName </th> <th> CountryCode </th>{' '}
          </tr>{' '}
        </thead>{' '}
        <tbody>
          {' '}
          {rows.map((row, index) => (
            <tr key={index}>
              <td> {row.IP} </td> <td> {row.CountryName} </td>{' '}
              <td> {row.CountryCode} </td>{' '}
            </tr>
          ))}{' '}
        </tbody>{' '}
      </table>{' '}
    </div>
  );
};

export default PaginatedTable;

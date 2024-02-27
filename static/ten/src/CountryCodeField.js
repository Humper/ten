import { hasFlag } from 'country-flag-icons';
import { useRecordContext } from 'react-admin';

const CountryCodeField = () => {
  const record = useRecordContext();
  return (
    <>
      {' '}
      {hasFlag(record.country_code) && (
        <img
          className="country-flag-icon"
          src={`http://purecatamphetamine.github.io/country-flag-icons/3x2/${record.country_code}.svg`}
          alt={record.country_code}
        />
      )}{' '}
      <span> {record.country_code} </span>{' '}
    </>
  );
};

export default CountryCodeField;

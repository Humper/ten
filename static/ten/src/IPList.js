import {
  Datagrid,
  List,
  TextField,
  SelectArrayInput,
  ReferenceInput,
  AutocompleteInput,
} from 'react-admin';
import IPDetailButtonField from './IPDetailButtonField';
import CountryCodeField from './CountryCodeField';

const countries = [
  'AT',
  'AU',
  'BE',
  'BG',
  'BR',
  'CA',
  'CH',
  'CL',
  'CR',
  'CY',
  'CZ',
  'DE',
  'DK',
  'EE',
  'EG',
  'ES',
  'ET',
  'FI',
  'FR',
  'GB',
  'GR',
  'HK',
  'HR',
  'HU',
  'ID',
  'IL',
  'IS',
  'IT',
  'JP',
  'KR',
  'KZ',
  'LI',
  'LT',
  'LU',
  'LV',
  'MD',
  'MX',
  'MY',
  'NL',
  'NO',
  'NZ',
  'PA',
  'PE',
  'PK',
  'PL',
  'PT',
  'RO',
  'RU',
  'SE',
  'SG',
  'SI',
  'TH',
  'TW',
  'UA',
  'US',
  'VN',
  'ZA',
];

const ipFilters = [
  <SelectArrayInput
    source="country_code"
    choices={countries.map((country) => ({ id: country, name: country }))}
  />,
];

export const IpList = () => (
  <List filters={ipFilters}>
    <Datagrid bulkActionButtons={false}>
      <TextField source="IP" />
      <TextField source="country_name" />
      <CountryCodeField source="country_code" />
      <IPDetailButtonField />
    </Datagrid>{' '}
  </List>
);

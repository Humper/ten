import { useRecordContext } from 'react-admin';

import IconButton from '@mui/material/IconButton';
import InfoIcon from '@mui/icons-material/Info';

const IPDetailButtonField = () => {
  const record = useRecordContext();
  return (
    <IconButton
      onClick={() => window.open(`https://iplocation.io/ip/${record.IP}`)}
    >
      <InfoIcon />
    </IconButton>
  );
};

export default IPDetailButtonField;

import PropTypes from 'prop-types';
import { TableCell, TableHead, TableRow } from '@mui/material';

const LogTableHead = ({ userIsAdmin }) => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>Time</TableCell>
        {userIsAdmin && <TableCell>Channels</TableCell>}
        {userIsAdmin && <TableCell>Users</TableCell>}
        <TableCell>Tokens</TableCell>
        <TableCell>Type</TableCell>
        <TableCell>模型</TableCell>
        <TableCell>提示</TableCell>
        <TableCell>补全</TableCell>
        <TableCell>Quota</TableCell>
        <TableCell>详情</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default LogTableHead;

LogTableHead.propTypes = {
  userIsAdmin: PropTypes.bool
};

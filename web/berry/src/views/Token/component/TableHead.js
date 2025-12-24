import { TableCell, TableHead, TableRow } from '@mui/material';

const TokenTableHead = () => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>Name</TableCell>
        <TableCell>Status</TableCell>
        <TableCell>已用额度</TableCell>
        <TableCell>剩余额度</TableCell>
        <TableCell>Created At</TableCell>
        <TableCell>Expires At</TableCell>
        <TableCell>操作</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default TokenTableHead;

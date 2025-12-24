import { TableCell, TableHead, TableRow } from '@mui/material';

const RedemptionTableHead = () => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>ID</TableCell>
        <TableCell>Name</TableCell>
        <TableCell>Status</TableCell>
        <TableCell>Quota</TableCell>
        <TableCell>Created At</TableCell>
        <TableCell>兑换时间</TableCell>
        <TableCell>操作</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default RedemptionTableHead;

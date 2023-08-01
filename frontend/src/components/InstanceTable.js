import { useNavigate } from "react-router-dom";
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import { Button } from '@mui/material';

const InstanceTable = (props) => {
    const navigate = useNavigate();
    const handleEditInstance = (row) => {
        navigate("/instance/edit", {state:row})
    }

    return (
        <TableContainer component={Paper}>
        <Table sx={{ minWidth: 650 }} aria-label="simple table">
            <TableHead>
            <TableRow>
                <TableCell><strong>Name</strong></TableCell>
                <TableCell align="left"><strong>Plan</strong></TableCell>
                <TableCell align="right"><strong>Actions</strong></TableCell>
            </TableRow>
            </TableHead>
            <TableBody>
            {props.instances.map((row) => (
                <TableRow
                key={row.name}
                sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                >
                <TableCell component="th" scope="row">
                    <Button variant="text">{row.name}</Button>
                </TableCell>
                <TableCell align="left">{row.plan.toUpperCase()}</TableCell>
                <TableCell align="right">
                    <Button onClick={handleEditInstance.bind(this,row)} size="small" variant="outlined">Edit</Button>
                </TableCell>
                </TableRow>
            ))}
            </TableBody>
        </Table>
        </TableContainer>
    );
}

export default InstanceTable;

import { DateInput, Create, NumberInput, SimpleForm, TextInput } from 'react-admin';

export const UserCreate = (props) => (
    <Create>
        <SimpleForm>
            <TextInput source="Name" />
            <TextInput source="Email" />
            <TextInput source="Password" />
            <TextInput source="Role" />
            <TextInput source="AllowedIPs" />
        </SimpleForm>
    </Create>
);
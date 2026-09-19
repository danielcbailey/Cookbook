import { useState, type CSSProperties } from "react";
import { Button, Checkbox, Form, Input, Label, Surface, TextField, Typography, toast } from "@heroui/react";
import { Header } from "../../header/header";
import { getPostLoginRedirect } from "../../helpers";
import { login, UserAPIError } from "../../userAPI";

const pageContainerStyle: CSSProperties = {
    height: '100vh',
    width: '100%',
};

const contentContainerStyle: CSSProperties = {
    height: '100%',
    width: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
};

const surfaceStyle: CSSProperties = {
    display: 'flex',
    flexDirection: 'column',
    gap: 10,
    borderRadius: 16,
    padding: 20,
    minWidth: 400,
};

/** Reads a text value out of the submitted form, ignoring anything that is not a string. */
function formString(data: FormData, name: string): string {
    const value = data.get(name);
    return typeof value === 'string' ? value : '';
}

export function LoginPage() {
    const [signingIn, setSigningIn] = useState(false);

    const onSubmit = (evt: React.FormEvent<HTMLFormElement>) => {
        evt.preventDefault();

        const data = new FormData(evt.currentTarget);
        const email = formString(data, 'email').trim();
        const password = formString(data, 'password');
        // An unchecked box is absent from the form data entirely.
        const longLived = data.get('remember-login') !== null;

        if (!email || !password) {
            toast.danger("Missing Credentials", {
                description: "Enter both your email address and your password.",
            });
            return;
        }

        setSigningIn(true);
        login(email, password, longLived).then(() => {
            // Navigated outside of the router so the app remounts and loads the
            // profile for the session that was just started.
            window.location.assign(getPostLoginRedirect());
        }).catch((reason: UserAPIError) => {
            setSigningIn(false);

            if (reason.status === 401) {
                toast.danger("Incorrect Email or Password", {
                    description: "Check your details and try again.",
                });
                return;
            }

            toast.danger("Failed to Sign In", {
                description: reason.message,
            });
        });
    }

    return (<div style={pageContainerStyle}>
        <Header loginVariant/>
        <div style={contentContainerStyle}>
            <Surface style={surfaceStyle}>
                <Typography type="h3" weight="bold" style={{minWidth: '100%', textAlign: 'center'}}>Log In</Typography>
                <Form style={{display: 'flex', gap: 10, flexDirection: 'column'}} onSubmit={onSubmit}>
                    <TextField fullWidth name="email" variant="secondary">
                        <Label>Email</Label>
                        <Input placeholder="email" type="email"/>
                    </TextField>
                    <TextField fullWidth name="password" variant="secondary">
                        <Label>Password</Label>
                        <Input placeholder="password" type="password"/>
                    </TextField>
                    <Checkbox name="remember-login" variant="secondary" style={{marginTop: 10, marginBottom: 10}}>
                        <Checkbox.Content>
                            <Checkbox.Control>
                                <Checkbox.Indicator/>
                            </Checkbox.Control>
                            Remember login details for next time
                        </Checkbox.Content>
                    </Checkbox>
                    <Button type="submit" isDisabled={signingIn}>{signingIn ? "Signing In..." : "Sign In"}</Button>
                </Form>
            </Surface>
        </div>
        
    </div>);
}
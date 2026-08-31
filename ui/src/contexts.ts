import { createContext } from "react";
import { type User } from "./apiTypes";


export const UserContext = createContext<User | null>(null);
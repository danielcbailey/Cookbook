import './App.css'
import { toast, Toast } from '@heroui/react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { PlanPage } from './pages/plan/page'
import { RecipesPage } from './pages/recipes/page'
import { PantryPage } from './pages/pantry/page'
import { ProfilePage } from './pages/profile/page'
import { RecipeViewPage } from './pages/recipe_view/page'
import { UserContext } from './contexts'
import type { User } from './apiTypes'
import { useEffect, useState } from 'react'
import { getCurrentUser, UserAPIError } from './userAPI'
import { RecipeEditPage } from './pages/recipe_edit/page'

function App() {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    getCurrentUser().then((user: User) => {
      setUser(user);
    }).catch((reason: UserAPIError) => {
      toast.danger("Failed to Retrieve Your Profile", {
          description: reason.message,
      });
    });
  }, []);

  return (
    <UserContext value={user}>
      <BrowserRouter>
        {/* Renders the toasts enqueued by the module-level toast() helpers. */}
        <Toast.Provider/>
        <Routes>
          <Route path="/" element={<Navigate to="/plan" replace/>}/>
          <Route path="/plan" element={<PlanPage/>}/>
          <Route path="/recipes" element={<RecipesPage/>}/>
          <Route path="/recipes/edit/:id" element={<RecipeEditPage/>}/>
          <Route path="/recipes/create" element={<RecipeEditPage/>}/>
          <Route path="/recipes/view/:id" element={<RecipeViewPage parent="Recipes"/>}/>
          <Route path="/plan/recipe/:id" element={<RecipeViewPage parent="Plan"/>}/>
          <Route path="/pantry" element={<PantryPage/>}/>
          <Route path="/profile" element={<ProfilePage/>}/>
        </Routes>
      </BrowserRouter>
    </UserContext>
  )
}

export default App

import { Link } from "react-router"

import { can } from "@/lib/auth"
import { PERMISSIONS } from "@/lib/constants"
import { useSession } from "@/lib/session"

export function CapacityFormulaNote() {
  const permissions = useSession((state) => state.permissions)
  const canManageHolidays = can(permissions, PERMISSIONS.departmentManage)

  return (
    <p className="text-xs text-muted-foreground">
      Effective capacity is hours per day times Monday–Friday, minus company holidays. Personal leave
      is not subtracted.
      {canManageHolidays ? (
        <>
          {" "}
          <Link className="underline-offset-4 hover:underline" to="/holidays">
            Manage holidays
          </Link>
        </>
      ) : null}
    </p>
  )
}

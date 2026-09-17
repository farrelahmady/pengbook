import AccountTopbar from "@/components/akun/account-topbar";
import AccountList from "@/components/akun/account-list";
import CreateAccount from "@/components/akun/create-account";

export default function AkunPage() {
	return (
		<>
			<AccountTopbar />
			<AccountList />
			<CreateAccount />
		</>
	);
}

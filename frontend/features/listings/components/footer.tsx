import Link from "next/link";

export function Footer() {
  return (
    <footer className="bg-[#181818] px-6 py-12 text-white sm:px-12">
      <div className="mx-auto flex max-w-[1580px] flex-col gap-16 lg:flex-row lg:justify-between">
        
        {/* Left Side: Newsletter & Addresses */}
        <div className="flex flex-col gap-10 lg:w-1/2">
          <div>
            <h3 className="mb-4 text-xl font-semibold">Subscribe to our Newsletter!</h3>
            <div className="flex max-w-sm items-center border-b border-gray-600 pb-2">
              <input 
                type="email" 
                placeholder="Email address" 
                className="w-full bg-transparent text-sm text-white outline-none placeholder:text-gray-400"
              />
              <button aria-label="Submit newsletter">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
              </button>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-8 text-xs text-gray-400 sm:grid-cols-3">
            <div>
              <p className="mb-2 font-semibold text-white">New York</p>
              <p>55 East 52nd Street, 23rd Floor,</p>
              <p>New York, NY 10022</p>
            </div>
            <div>
              <p className="mb-2 font-semibold text-white">Emails</p>
              <p>hello@findrealestate.com</p>
            </div>
            <div>
              <p className="mb-2 font-semibold text-white">Calls</p>
              <p>+1202 555 0122</p>
            </div>
            <div>
              <p className="mb-2 pt-4 font-semibold text-white">Miami</p>
              <p>55 East 52nd Street, 23rd Floor,</p>
              <p>Miami, FL 33132</p>
            </div>
            <div>
              <p className="mb-2 pt-4 font-semibold text-white">Emails</p>
              <p>hello@findrealestate.com</p>
            </div>
            <div>
              <p className="mb-2 pt-4 font-semibold text-white">Calls</p>
              <p>+1202 555 0122</p>
            </div>
          </div>
        </div>

        {/* Right Side: Links */}
        <div className="flex lg:justify-end">
          <div className="grid grid-cols-2 gap-x-20 gap-y-3 text-sm font-medium">
            <Link href="/search" className="hover:text-gray-300">Search</Link>
            <Link href="/instagram" className="hover:text-gray-300">Instagram</Link>
            
            <Link href="/agents" className="hover:text-gray-300">Agents</Link>
            <Link href="/youtube" className="hover:text-gray-300">Youtube</Link>
            
            <Link href="/join" className="hover:text-gray-300">Join</Link>
            <Link href="/linkedin" className="hover:text-gray-300">LinkedIn</Link>
            
            <Link href="/about" className="hover:text-gray-300">About Us</Link>
            <div />
            
            <Link href="/agent-portal" className="hover:text-gray-300">Agent Portal</Link>
            <div />
          </div>
        </div>
      </div>

      {/* Bottom Area: Massive Logo and Copyright */}
      <div className="mx-auto mt-20 max-w-[1580px] border-t border-gray-800 pt-8">
        <h1 className="text-[12vw] font-black leading-none tracking-tighter sm:text-[10rem]">
          FIND
        </h1>
        <div className="mt-8 flex flex-wrap items-center justify-between gap-4 text-[10px] text-gray-500 uppercase tracking-wider">
          <div className="flex gap-4">
            <Link href="/terms" className="hover:text-white">Terms</Link>
            <Link href="/privacy" className="hover:text-white">Privacy policy</Link>
            <Link href="/fair-housing" className="hover:text-white">Fair Housing Notice</Link>
            <Link href="/terms-of-use" className="hover:text-white">Terms of Use</Link>
            <Link href="/press" className="hover:text-white">Press</Link>
            <Link href="/do-not-sell" className="hover:text-white">Do Not Sell or Share My Personal Information</Link>
          </div>
          <div>
            Copyright © 2024 FIND Real Estate
          </div>
        </div>
      </div>
    </footer>
  );
}

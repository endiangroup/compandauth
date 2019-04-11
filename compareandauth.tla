--------------------------- MODULE compareandauth ---------------------------
EXTENDS Integers, TLC
CONSTANTS DELTA, MAX_SESSIONS

(* --algorithm compareandauth
variables master_caa = 0, issued_sessions = {}

define
isPositive(x) == x >= 0
abs(x) == IF isPositive(x)
          THEN x
          ELSE -1 * x
step(x, y) == IF isPositive(x)
              THEN x + y
              ELSE x - y
step1(x) == step(x, 1)
max(j, k) == IF j >= k
             THEN j
             ELSE k
min(j, k) == IF j <= k
             THEN j
             ELSE k
isLocked(x) == x < 0
hasIssued(x) == x /= 0
isValid(x) == /\ ~isLocked(master_caa)
              /\ hasIssued(master_caa)
              /\ x + DELTA >= abs(master_caa) 
              /\ x < abs(master_caa)

MasterCaaValueNeverIssued == master_caa \notin issued_sessions
IssuedSessionsAlwaysPositiveIntegers == \A x \in issued_sessions: isPositive(x)
LastDeltaIssuedSessionsAreValid == hasIssued(master_caa) => {x \in issued_sessions: isValid(x)} \subseteq max(0, abs(master_caa)-DELTA) .. max(0, abs(master_caa)-1)
AllOtherIssuedSessionsAreInvalid == master_caa > DELTA => {x \in issued_sessions: ~isValid(x)} \subseteq 0 .. abs(master_caa)-DELTA-1
AllIssuedSessionsLessThanMasterCaa == hasIssued(master_caa) => issued_sessions \subseteq 0 .. abs(master_caa)-1
WhenLockedAllIssuedSessionsAreInvalid == hasIssued(master_caa) /\ isLocked(master_caa) => \A x \in issued_sessions: ~isValid(x)
AnySessionGreaterThanMasterCaaIsInvalid == \A x \in abs(master_caa) .. MAX_SESSIONS: ~isValid(x)
end define;

begin
Act:
    while abs(master_caa) < MAX_SESSIONS do
        either
            Issue: issued_sessions := issued_sessions \union {abs(master_caa)};
                   master_caa := step1(master_caa);
        or
            await ~isLocked(master_caa);
            Lock: master_caa := -1 * abs(master_caa);
        or
            await isLocked(master_caa);
            Unlock: master_caa := abs(master_caa);
        or
            await hasIssued(master_caa);
            Revoke: master_caa := step(master_caa, min(abs(master_caa), DELTA));
        end either; 
    end while;
end algorithm; *)

\* BEGIN TRANSLATION
VARIABLES master_caa, issued_sessions, pc

(* define statement *)
isPositive(x) == x >= 0
abs(x) == IF isPositive(x)
          THEN x
          ELSE -1 * x
step(x, y) == IF isPositive(x)
              THEN x + y
              ELSE x - y
step1(x) == step(x, 1)
max(j, k) == IF j >= k
             THEN j
             ELSE k
min(j, k) == IF j <= k
             THEN j
             ELSE k
isLocked(x) == x < 0
hasIssued(x) == x /= 0
isValid(x) == /\ ~isLocked(master_caa)
              /\ hasIssued(master_caa)
              /\ x + DELTA >= abs(master_caa)
              /\ x < abs(master_caa)

MasterCaaValueNeverIssued == master_caa \notin issued_sessions
IssuedSessionsAlwaysPositiveIntegers == \A x \in issued_sessions: isPositive(x)
LastDeltaIssuedSessionsAreValid == hasIssued(master_caa) => {x \in issued_sessions: isValid(x)} \subseteq max(0, abs(master_caa)-DELTA) .. max(0, abs(master_caa)-1)
AllOtherIssuedSessionsAreInvalid == master_caa > DELTA => {x \in issued_sessions: ~isValid(x)} \subseteq 0 .. abs(master_caa)-DELTA-1
AllIssuedSessionsLessThanMasterCaa == hasIssued(master_caa) => issued_sessions \subseteq 0 .. abs(master_caa)-1
WhenLockedAllIssuedSessionsAreInvalid == hasIssued(master_caa) /\ isLocked(master_caa) => \A x \in issued_sessions: ~isValid(x)
AnySessionGreaterThanMasterCaaIsInvalid == \A x \in abs(master_caa) .. MAX_SESSIONS: ~isValid(x)


vars == << master_caa, issued_sessions, pc >>

Init == (* Global variables *)
        /\ master_caa = 0
        /\ issued_sessions = {}
        /\ pc = "Act"

Act == /\ pc = "Act"
       /\ IF abs(master_caa) < MAX_SESSIONS
             THEN /\ \/ /\ pc' = "Issue"
                     \/ /\ ~isLocked(master_caa)
                        /\ pc' = "Lock"
                     \/ /\ isLocked(master_caa)
                        /\ pc' = "Unlock"
                     \/ /\ hasIssued(master_caa)
                        /\ pc' = "Revoke"
             ELSE /\ pc' = "Done"
       /\ UNCHANGED << master_caa, issued_sessions >>

Issue == /\ pc = "Issue"
         /\ issued_sessions' = (issued_sessions \union {abs(master_caa)})
         /\ master_caa' = step1(master_caa)
         /\ pc' = "Act"

Lock == /\ pc = "Lock"
        /\ master_caa' = -1 * abs(master_caa)
        /\ pc' = "Act"
        /\ UNCHANGED issued_sessions

Unlock == /\ pc = "Unlock"
          /\ master_caa' = abs(master_caa)
          /\ pc' = "Act"
          /\ UNCHANGED issued_sessions

Revoke == /\ pc = "Revoke"
          /\ master_caa' = step(master_caa, min(abs(master_caa), DELTA))
          /\ pc' = "Act"
          /\ UNCHANGED issued_sessions

Next == Act \/ Issue \/ Lock \/ Unlock \/ Revoke
           \/ (* Disjunct to prevent deadlock on termination *)
              (pc = "Done" /\ UNCHANGED vars)

Spec == Init /\ [][Next]_vars

Termination == <>(pc = "Done")

\* END TRANSLATION
=============================================================================
\* Modification History
\* Last modified Fri Mar 15 13:53:53 GMT 2019 by adrianduke
\* Created Sat Aug 12 16:10:55 BST 2017 by adrianduke

(*

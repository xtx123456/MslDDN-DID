import math
import scipy.special
from fractions import Fraction
from tabulate import tabulate
# For all functions we use
#n=total population
#number of corruptions in total population
#k=security parameter
# Probability of having < h honest parties in committee
# s = committee size
# h m minimum number of honest parties required in committee
# p_max = maximal value for which pFail returns correct value, 0utput is 1 if p_fail >p_max.
def p_fail(n, t , s , h, p_max):
    p=0
    # compute n choose s as exact integer
    denom = scipy .special.comb(n, s , exact=True)
    for i in range(s - h + 1, s+1) :
        p += Fraction(scipy.special.comb(t,i, exact=True) * scipy.special.comb(n-t,s-i,exact=True),denom)
    if p > p_max :
        return 1
    return p

# Minimum committee size with corruption ratio at most cr such that pfail <= 2~(-k)
def min_csize(n, t , cr , k) :
    p_max = Fraction(1, 2**k)
    for s in range(1, n+1) :
        #we want at least h honest parties
        h = math.ceil((1 - cr) * s)
        if p_fail(n,t,s, h, p_max) <= p_max :
            return s

# Compute analytical upper bound
# p = fraction of corruption in total population
# cr = maximal corruption ratio in committee
def analytic_bound(p, cr , k) :
    q = cr
    alpha = q - p
    beta = (math.e * math.sqrt(q) * (1 - p)) / (2 * math.pi * alpha * math.sqrt(1 - q))
    bound = math.ceil(k / math.log((q / p) ** q * ((1 - q) / (1 - p)) ** (1 - q), 2))
    beta_bound = math.ceil(beta * beta)
    return max(bound, beta,bound)

# Print committee sizes for different parameters
# ns = list of values of n,total population
# ps = list of corruption fractions in total population# k = security parameter
# crs = list of maximal corruption ratios in committee
def print_csizes (ns , ps ,k ,crs):
    header = ["",""] + [float(cr) for cr in crs]
    table = []
    for p in ps:
        table += [["analytics","p = "+str(p)] + [analytic_bound(p, cr, k) for cr in crs]]
        for n in ns:
            t=n*p
            table +=[ ["n ="+ str(n), "p = " + str(p)] + [min_csize(n, t, cr, k) for cr in crs]]
    print(tabulate(table , header))

# Print coordinates of committee sizes over guaranteed honesty
def print_graph_coordinates ( n , t , k ) :
    for crp in range ( 99 , 32 ,-1) :   # corruption from 99% down to 33%
        cr = Fraction ( crp , 100) # convert percentage to fraction
    print ( "(" , int ( 100 - 100*cr) , "," , min_csize ( n , t , cr , k ) , ")" ,sep='', end=' ' )
print ( "\n" )

# Print analytic bound for coordinates of committee sizes over guaranteed honesty
# p = fraction of corruption in total population
def print_analytic_graph_coordinates ( p , k ) :
    for crp in range ( 99 , 32 , -1) : # corruption from 99% down to 33%
        cr = Fraction ( crp , 100 ) # convert percentage to fraction
        print ( "(" , int ( 100 - 100*cr) , "," , analytic_bound ( p , cr , k ) , ")" , sep='', end=' ' )
print ( "\n" )

def main ( ) :
# Consider total populations 1000 and 500 with 30% and 20% corruption , with 60 bit security
    ns = [ 1000 , 500]
    ps = [ Fraction ( 3 , 10) , Fraction ( 2 , 10)]
    k = 60

    # maximal corruption ratios in committees from 99% to 33%
    crs = [ Fraction ( 90 , 100 ) , Fraction ( 80 , 100 ) , Fraction ( 70 , 100 ) , Fraction ( 60 , 100) ,
    Fraction ( 50, 100 ), Fraction ( 40 , 100),Fraction ( 1 , 3 )]

    # print table with committee sizes
    print_csizes ( ns , ps , k , crs)
    print ( "\n" )


if __name__ == '__main__':
    main()

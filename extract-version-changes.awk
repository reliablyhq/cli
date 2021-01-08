#!/usr/bin/awk -f


BEGIN {
  found = 0;
  RS="\n"
  FS=" "

  versiontag="[" v "]"
  versionlink="[" v "]:"

}

# Beware the rules orders !!

# 1 -> We find the changes section that matches the version
$1 == "##" && $2 == versiontag {
  print;
  found=1;
  next
}


# 2 -> Once changes section is found, next time we encounter ##
# that 's another changes section, so we need to deactivate found not to print anymore
$1 == "##" && found == 1 {
  found=0;
  next
}


# 3 -> When found is True, we are within the targeted changes section, print full line
found{
  print;
  next
};

# 4 -> When found is not true, we still need to find the trailing version link
!found && $1 == versionlink{
  print;
  next
}

END {

}